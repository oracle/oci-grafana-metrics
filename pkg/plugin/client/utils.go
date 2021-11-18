package client

import (
	"bufio"
	"context"
	"errors"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	jsoniter "github.com/json-iterator/go"
	"github.com/oracle/oci-go-sdk/v51/common"
	"github.com/oracle/oci-go-sdk/v51/common/auth"
	"github.com/oracle/oci-go-sdk/v51/core"
	"github.com/oracle/oci-go-sdk/v51/identity"
	"github.com/oracle/oci-go-sdk/v51/monitoring"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

// New Creates a new OCI client from the DatasourceInfo
func newClientPerProfile(authProvider string, configPath string, configProfile string) (*OCIClient, error) {
	backend.Logger.Info("client.utils", "newClientPerProfile", configProfile)

	ociClient := OCIClient{}
	var err error

	cp := getOCIConfigurationProvider(authProvider, configPath, configProfile)
	if cp == nil {
		backend.Logger.Error("client.utils", "newClientPerProfile", "OCI provider is not working properly, please check IAM policy")
		return nil, errors.New("oci provider is not working properly, please check")
	}
	ociClient.clientConfigProvider = cp

	// setting tenancy ocid
	ociClient.tenancyOCID, err = cp.TenancyOCID()
	if err != nil {
		backend.Logger.Error("client.utils", "newClientPerProfile", "could not fetch base tenancy ocid: "+err.Error())
		return nil, err
	}

	// setting base region
	ociClient.region, err = cp.Region()
	if err != nil {
		backend.Logger.Error("client.utils", "newClientPerProfile", "could not fetch base region: "+err.Error())
		return nil, err
	}

	irp := clientRetryPolicy()
	// creating oci identity client
	ociClient.identityClient, err = identity.NewIdentityClientWithConfigurationProvider(cp)
	if err != nil {
		backend.Logger.Error("client.utils", "newClientPerProfile", "could not create oci identity client: "+err.Error())
		return nil, err
	}
	ociClient.identityClient.Configuration.RetryPolicy = &irp

	mrp := clientRetryPolicy()
	// creating oci monitoring client
	ociClient.monitoringClient, err = monitoring.NewMonitoringClientWithConfigurationProvider(cp)
	if err != nil {
		backend.Logger.Error("client.utils", "newClientPerProfile", "could not create oci monitoring client: "+err.Error())
		return nil, err
	}
	ociClient.monitoringClient.Configuration.RetryPolicy = &mrp

	crp := clientRetryPolicy()
	// creating oci core compute client
	ociClient.computeClient, err = core.NewComputeClientWithConfigurationProvider(cp)
	if err != nil {
		backend.Logger.Error("client.utils", "newClientPerProfile", "could not create oci core compute client: "+err.Error())
		return nil, err
	}
	ociClient.computeClient.Configuration.RetryPolicy = &crp

	return &ociClient, nil
}

// clientRetryPolicy is a helper method that assembles and returns a return policy that is defined to call in every second
// to use maximum benefit of TPS limit (which is currently 10)
// This retry policy will retry on (409, IncorrectState), (429, TooManyRequests) and any 5XX errors except (501, MethodNotImplemented)
// The retry behavior is constant with 1s
// The number of retries is 10
func clientRetryPolicy() common.RetryPolicy {
	clientRetryOperation := func(r common.OCIOperationResponse) bool {
		type HTTPStatus struct {
			code    int
			message string
		}
		clientRetryStatusCodeMap := map[HTTPStatus]bool{
			{409, "IncorrectState"}:       true,
			{429, "TooManyRequests"}:      true,
			{501, "MethodNotImplemented"}: false,
		}

		if r.Error == nil && 199 < r.Response.HTTPResponse().StatusCode && r.Response.HTTPResponse().StatusCode < 300 {
			return false
		}
		if common.IsNetworkError(r.Error) {
			return true
		}
		if err, ok := common.IsServiceError(r.Error); ok {
			if shouldRetry, ok := clientRetryStatusCodeMap[HTTPStatus{err.GetHTTPStatusCode(), err.GetCode()}]; ok {
				return shouldRetry
			}
			return 500 <= r.Response.HTTPResponse().StatusCode && r.Response.HTTPResponse().StatusCode < 600
		}
		return false
	}
	nextCallAt := func(r common.OCIOperationResponse) time.Duration {
		return time.Duration(1) * time.Second
	}
	return common.NewRetryPolicy(uint(10), clientRetryOperation, nextCallAt)
}

// getOCIConfigurationProvider Creates oci configuration provider based on the provider
func getOCIConfigurationProvider(authProvider string, configPath string, configProfile string) common.ConfigurationProvider {
	if authProvider == constants.OCI_INSTANCE_AUTH_PROVIDER {
		cp, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil
		}

		return cp
	}

	customFileProvider, _ := common.ConfigurationProviderFromFileWithProfile(configPath, configProfile, "")
	configProvider, _ := common.ComposingConfigurationProvider([]common.ConfigurationProvider{customFileProvider})
	return configProvider
}

// readMultiTenancySourceFile Reads either default or user provided source file for all remote tenancies
func readMultiTenancySourceFile(filePath string, tenanciesMap map[string]string) error {
	backend.Logger.Debug("client.utils", "readMultiTenancySourceFile", "reading tenancies from file:"+filePath)

	f, err := os.Open(filePath)
	if err != nil {
		backend.Logger.Error("client.utils", "readMultiTenancySourceFile", "could not open Multi-Tenancy File: "+err.Error())
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			backend.Logger.Error("client.utils", "readMultiTenancySourceFile", "could not close Multi-Tenancy File: "+err.Error())
			return
		}
	}()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		// when there is blank line
		if line == "" {
			continue
		}

		// creating tenancies map, key=tenancy_ocid, value=tenancy_name
		t := strings.Split(line, ",")
		// line has no proper content
		if len(t) != 2 {
			continue
		}
		tenanciesMap[t[1]] = t[0]
	}
	err = s.Err()
	if err != nil {
		backend.Logger.Error("client.utils", "readMultiTenancySourceFile", "could not read Multi-Tenancy File: "+err.Error())
		return err
	}

	return nil
}

/*
func constructCMDBData(cmdbFileData string) (map[string]models.CMDBCustomerData, error) {
	backend.Logger.Debug("client.utils", "constructCMDBData", "converting data in proper format")

	var formattedCMDBFileData map[string][]models.CMDBFileData
	if err := jsoniter.Unmarshal([]byte(cmdbFileData), &formattedCMDBFileData); err != nil {
		backend.Logger.Error("client.utils", "constructCMDBData", "converting data in proper format: "+err.Error())
		return nil, err
	}

	cmdbData := map[string]models.CMDBCustomerData{}
	for tenancyName, tenancyResourceData := range formattedCMDBFileData {
		tenancyCMDBData := models.CMDBCustomerData{}
		cmdbEnvData := map[string]map[string][]string{}

		for _, resourceData := range tenancyResourceData {
			if tenancyCMDBData.Customer == "" {
				tenancyCMDBData.Customer = resourceData.Customer
			}

			// for new environment entry
			if _, ok := cmdbEnvData[resourceData.EnvironmentName]; !ok {
				cmdbEnvData[resourceData.EnvironmentName] = map[string][]string{
					resourceData.ResourceType: {resourceData.ResourceOCID},
				}
				continue
			}

			// for new resource type entry under environment
			if _, ok := cmdbEnvData[resourceData.EnvironmentName][resourceData.ResourceType]; !ok {
				cmdbEnvData[resourceData.EnvironmentName][resourceData.ResourceType] = []string{resourceData.ResourceOCID}
				continue
			}

			cmdbEnvData[resourceData.EnvironmentName][resourceData.ResourceType] = append(cmdbEnvData[resourceData.EnvironmentName][resourceData.ResourceType], resourceData.ResourceOCID)
		}

		tenancyCMDBData.EnvironmentData = cmdbEnvData
		cmdbData[tenancyName] = tenancyCMDBData
	}

	backend.Logger.Warn("client.utils", "constructCMDBData", cmdbData)

	return cmdbData, nil
}
*/

// constructCMDBData will format uploaded cmdb file data as per requirement to show as labels
func constructCMDBData(cmdbFileData string) (map[string]map[string]map[string]string, error) {
	backend.Logger.Debug("client.utils", "constructCMDBData", "converting data in proper format")

	var formattedCMDBFileData map[string][]models.CMDBFileData
	if err := jsoniter.Unmarshal([]byte(cmdbFileData), &formattedCMDBFileData); err != nil {
		backend.Logger.Error("client.utils", "constructCMDBData", "converting data in proper format: "+err.Error())
		return nil, err
	}

	// to store data in the following format:
	// tenancyName: ocid: {labels}
	cmdbData := map[string]map[string]map[string]string{}
	for tenancyName, tenancyResourceData := range formattedCMDBFileData {
		tenancyCMDBData := map[string]map[string]string{}

		for _, resourceData := range tenancyResourceData {
			tenancyCMDBData[resourceData.ResourceOCID] = map[string]string{
				"customer":      resourceData.Customer,
				"environment":   resourceData.EnvironmentName,
				"resource_type": resourceData.ResourceType,
			}
		}

		cmdbData[tenancyName] = tenancyCMDBData
	}

	return cmdbData, nil
}

// listMetrics will list all metrics with namespaces
// API Operation: ListMetrics
// Permission Required: METRIC_INSPECT
// Links:
// https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/monitoringpolicyreference.htm
// https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/Metric/ListMetrics
func listMetrics(ctx context.Context, mClient monitoring.MonitoringClient, req monitoring.ListMetricsRequest) []monitoring.Metric {
	var fetchedMetricDetails []monitoring.Metric
	var pageHeader string

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		res, err := mClient.ListMetrics(ctx, req)
		if err != nil {
			backend.Logger.Error("client.utils", "listMetrics", err)
			break
		}

		fetchedMetricDetails = append(fetchedMetricDetails, res.Items...)
		if len(res.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *res.OpcNextPage
		} else {
			break
		}
	}

	return fetchedMetricDetails
}

// listMetricsMetadataPerRegion will list all either dimensions or metrics either per namespaces or per resurce groups for a particular region
func listMetricsMetadataPerRegion(
	ctx context.Context,
	ci *ristretto.Cache,
	cacheKey string,
	fetchFor string,
	mClient monitoring.MonitoringClient,
	req monitoring.ListMetricsRequest) map[string][]string {

	backend.Logger.Debug("client.utils", "listMetricsMetadataPerRegion", "Data fetch start by calling list metrics API for a particular regions")

	if cachedMetricsData, found := ci.Get(cacheKey); found {
		backend.Logger.Warn("client.utils", "listMetricsMetadataPerRegion", "getting the data from cache -> "+cacheKey)
		return cachedMetricsData.(map[string][]string)
	}

	fetchedMetricDetails := listMetrics(ctx, mClient, req)

	metadataWithMetricNames := map[string][]string{}
	sortedMetadataWithMetricNames := map[string][]string{}
	metadata := []string{}
	isExist := false
	var metadataKey string

	for _, item := range fetchedMetricDetails {
		metricName := *item.Name

		switch fetchFor {
		case constants.FETCH_FOR_NAMESPACE:
			metadataKey = *item.Namespace
		case constants.FETCH_FOR_RESOURCE_GROUP:
			if item.ResourceGroup != nil {
				metadataKey = *item.ResourceGroup
			}
		case constants.FETCH_FOR_DIMENSION:
			for k, v := range item.Dimensions {
				// we don't need region or resource id dimensions as
				// we already filtered by region and resourceDisplayName is already there
				// in the dimensions
				// and do we really need imageId, image name will be good to have
				if k == "region" || k == "resourceId" || k == "imageId" {
					continue
				}

				// to sort the final map by dimension keys
				metadata = append(metadata, k)

				if oldVs, ok := metadataWithMetricNames[k]; ok {
					for _, oldv := range oldVs {
						if v == oldv {
							isExist = true
							break
						}
					}
					if !isExist {
						metadataWithMetricNames[k] = append(metadataWithMetricNames[k], v)
					}

					isExist = false
					continue
				}

				metadataWithMetricNames[k] = []string{v}
			}
		}

		if len(metadataKey) == 0 {
			// in case of dimensions
			continue
		}

		// to sort the final map by namespaces
		metadata = append(metadata, metadataKey)

		if _, ok := metadataWithMetricNames[metadataKey]; ok {
			metadataWithMetricNames[metadataKey] = append(metadataWithMetricNames[metadataKey], metricName)
			continue
		}

		metadataWithMetricNames[metadataKey] = []string{metricName}
	}

	// sorting the metadata key values
	if len(metadata) != 0 {
		sort.Strings(metadata)
	}

	// generating new map with sorted metadata keys and metric names
	for _, md := range metadata {
		sort.Strings(metadataWithMetricNames[md])
		sortedMetadataWithMetricNames[md] = metadataWithMetricNames[md]
	}

	ci.SetWithTTL(cacheKey, sortedMetadataWithMetricNames, 1, 5*time.Minute)
	ci.Wait()

	return sortedMetadataWithMetricNames
}

// listMetricsMetadataFromAllRegion will list either metric names with namespaces or dimensions for all subscribed region
func listMetricsMetadataFromAllRegion(
	ctx context.Context,
	ci *ristretto.Cache,
	cacheKey string,
	fetchFor string,
	mClient monitoring.MonitoringClient,
	req monitoring.ListMetricsRequest,
	regions []string) map[string][]string {

	backend.Logger.Debug("client.utils", "listMetricsMetadataFromAllRegion", "Data fetch start by calling list metrics API from all subscribed regions")

	var metricsMetadata map[string][]string
	var allRegionsData sync.Map
	var wg sync.WaitGroup

	for _, subscribedRegion := range regions {
		if subscribedRegion != constants.ALL_REGION {
			mClient.SetRegion(subscribedRegion)

			wg.Add(1)
			go func(mc monitoring.MonitoringClient, sRegion string) {
				defer wg.Done()

				newCacheKey := strings.ReplaceAll(cacheKey, constants.ALL_REGION, subscribedRegion)
				metadata := listMetricsMetadataPerRegion(ctx, ci, newCacheKey, fetchFor, mc, req)

				if len(metadata) > 0 {
					allRegionsData.Store(sRegion, metadata)
				}
			}(mClient, subscribedRegion)
		}
	}
	wg.Wait()

	allRegionsData.Range(func(key, value interface{}) bool {
		backend.Logger.Info("client.utils", "listMetricsMetadataPerAllRegion", "Data got for region-"+key.(string))

		metadataGot := value.(map[string][]string)

		// for first entry
		if len(metricsMetadata) == 0 {
			metricsMetadata = metadataGot
			return true
		}

		// k can be namespace or dimension key
		// values can be either metricNames or dimension values
		for k, values := range metadataGot {
			if _, ok := metricsMetadata[k]; !ok {
				// when namespace not present
				metricsMetadata[k] = values
				continue
			}

			// when namespace is already present
			for _, mn := range values {
				findIndex := sort.SearchStrings(metricsMetadata[k], mn)
				if findIndex < len(metricsMetadata[k]) && metricsMetadata[k][findIndex] != mn {
					// not found, and insert in between
					metricsMetadata[k] = append(metricsMetadata[k][:findIndex+1], metricsMetadata[k][findIndex:]...)
					metricsMetadata[k][findIndex] = mn
				} else if findIndex == len(metricsMetadata[k]) {
					// not found and insert at last
					metricsMetadata[k] = append(metricsMetadata[k], mn)
				}
			}
		}

		return true
	})

	return metricsMetadata
}

func fetchResourceTags(resourceTagsResponse []models.OCIResourceTagsResponse) (map[string][]string, map[string]map[string]struct{}) {
	backend.Logger.Debug("client.utils", "fetchResourceTags", "Fetching the tags from the oci call response")

	// holds key: value1, value2, for UI
	resourceTags := map[string][]string{}
	// holds key.value: map of resourceIDs, for caching
	resourceIDsPerTag := map[string]map[string]struct{}{}

	for _, item := range resourceTagsResponse {
		resourceID := item.ResourceID
		isExist := false
		// for defined tags
		for rootTagKey, rootTags := range item.DefinedTags {
			if rootTagKey == "Oracle-Tags" {
				continue
			}

			for k, v := range rootTags {
				if k == "Created_On" {
					continue
				}
				tagValue := v.(string)

				key := strings.Join([]string{rootTagKey, k}, ".")
				cacheKey := strings.Join([]string{rootTagKey, k, tagValue}, ".")

				// for UI
				existingVs, ok := resourceTags[key]
				if ok {
					for _, oldv := range existingVs {
						if oldv == tagValue {
							isExist = true
							break
						}
					}
					if !isExist {
						resourceTags[key] = append(resourceTags[key], tagValue)
					}

					isExist = false
				} else {
					resourceTags[key] = []string{tagValue}
				}

				// for caching
				if len(resourceIDsPerTag[cacheKey]) == 0 {
					resourceIDsPerTag[cacheKey] = map[string]struct{}{
						resourceID: {},
					}
				} else {
					resourceIDsPerTag[cacheKey][resourceID] = struct{}{}
				}
			}
		}

		isExist = false
		// for freeform tags
		for k, v := range item.FreeFormTags {
			cacheKey := strings.Join([]string{k, v}, ".")

			// for UI
			existingVs, ok := resourceTags[k]
			if ok {
				for _, oldv := range existingVs {
					if v == oldv {
						isExist = true
						break
					}
				}
				if !isExist {
					resourceTags[k] = append(resourceTags[k], v)
				}

				isExist = false
			} else {
				resourceTags[k] = []string{v}
			}

			// for caching
			if len(resourceIDsPerTag[cacheKey]) == 0 {
				resourceIDsPerTag[cacheKey] = map[string]struct{}{
					resourceID: {},
				}
			} else {
				resourceIDsPerTag[cacheKey][resourceID] = struct{}{}
			}
		}
	}

	return resourceTags, resourceIDsPerTag
}

func collectResourceTags(resourceTagsResponse []models.OCIResourceTagsResponse) (map[string]map[string]struct{}, map[string]map[string]struct{}) {
	backend.Logger.Debug("client.utils", "collectResourceTags", "Fetching the tags from the oci call response")

	// holds key: map pf values, for UI
	resourceTags := map[string]map[string]struct{}{}
	// holds key.value: map of resourceIDs, for caching
	resourceIDsPerTag := map[string]map[string]struct{}{}

	for _, item := range resourceTagsResponse {
		resourceID := item.ResourceID
		// for defined tags
		for rootTagKey, rootTags := range item.DefinedTags {
			if rootTagKey == "Oracle-Tags" {
				continue
			}

			for k, v := range rootTags {
				if k == "Created_On" {
					continue
				}
				tagValue := v.(string)

				tagKey := strings.Join([]string{rootTagKey, k}, ".")
				cacheKey := strings.Join([]string{rootTagKey, k, tagValue}, ".")

				// for UI
				// when the tag key is already present
				if existingVs, ok := resourceTags[tagKey]; ok {
					// if the value for the tag key is not added before
					if _, found := existingVs[tagValue]; !found {
						existingVs[tagValue] = struct{}{}
						resourceTags[tagKey] = existingVs
					}
				} else {
					// when the tag key is added for first time
					resourceTags[tagKey] = map[string]struct{}{
						tagValue: {},
					}
				}

				// for caching
				if len(resourceIDsPerTag[cacheKey]) == 0 {
					resourceIDsPerTag[cacheKey] = map[string]struct{}{
						resourceID: {},
					}
				} else {
					resourceIDsPerTag[cacheKey][resourceID] = struct{}{}
				}
			}
		}

		// for freeform tags
		for tagKey, tagValue := range item.FreeFormTags {
			cacheKey := strings.Join([]string{tagKey, tagValue}, ".")

			// for UI
			// when the tag key is already present
			if existingVs, ok := resourceTags[tagKey]; ok {
				// if the value for the tag key is not added before
				if _, found := existingVs[tagValue]; !found {
					existingVs[tagValue] = struct{}{}
					resourceTags[tagKey] = existingVs
				}
			} else {
				// when the tag key is added for first time
				resourceTags[tagKey] = map[string]struct{}{
					tagValue: {},
				}
			}

			// for caching
			if len(resourceIDsPerTag[cacheKey]) == 0 {
				resourceIDsPerTag[cacheKey] = map[string]struct{}{
					resourceID: {},
				}
			} else {
				resourceIDsPerTag[cacheKey][resourceID] = struct{}{}
			}
		}
	}

	return resourceTags, resourceIDsPerTag
}

func convertToArray(input map[string]map[string]struct{}) map[string][]string {
	backend.Logger.Debug("client.utils", "convertToArray", "Converting to array")

	output := map[string][]string{}

	for key, values := range input {
		for v := range values {
			if len(output[key]) == 0 {
				output[key] = []string{v}
			} else {
				output[key] = append(output[key], v)
			}
		}
	}

	return output
}

func getUniqueIdsForLabels(namespace string, dimensions map[string]string) (string, string, string, bool) {
	monitorID := ""

	// getting the resource unique ID
	resourceID, found := dimensions["resourceId"]
	if !found {
		resourceID, found = dimensions["ResourceId"]
		if !found {
			// as only one key and value pair will be present based on group by key selection
			for _, v := range dimensions {
				resourceID = v
			}
		}
	}

	// getting the resource name
	resourceDisplayName := resourceID
	if v, got := dimensions["resourceDisplayName"]; got {
		resourceDisplayName = v
	}

	// getting the extra unique id as per namespace
	if namespace == constants.OCI_NS_APM {
		monitorID = dimensions["MonitorId"]
	}

	return resourceID, resourceDisplayName, monitorID, found
}

func addDimensionsAsLabels(namespace string, existingLabels map[string]string, dimensions map[string]string) map[string]string {
	if namespace != constants.OCI_NS_APM {
		return existingLabels
	}

	keysToInclude := map[string]struct{}{
		"ErrorCategory":    {},
		"Genre":            {},
		"OracleApmType":    {},
		"UserAgent":        {},
		"VantagePoint":     {},
		"VantagePointType": {},
	}

	for k, v := range dimensions {
		if _, ok := keysToInclude[k]; ok {
			existingLabels[strings.ToLower(k)] = v
		}
	}

	return existingLabels
}
