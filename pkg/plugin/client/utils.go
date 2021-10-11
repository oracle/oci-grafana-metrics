package client

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/common/auth"
	"github.com/oracle/oci-go-sdk/v49/core"
	"github.com/oracle/oci-go-sdk/v49/healthchecks"
	"github.com/oracle/oci-go-sdk/v49/identity"
	"github.com/oracle/oci-go-sdk/v49/loadbalancer"
	"github.com/oracle/oci-go-sdk/v49/monitoring"

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
	fptr := flag.String("tfpath", filePath, "tenancies file path to read from")
	flag.Parse()

	f, err := os.Open(*fptr)
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
		t := strings.Split(s.Text(), ",")
		tenanciesMap[t[1]] = t[0]
	}
	err = s.Err()
	if err != nil {
		backend.Logger.Error("client.utils", "readMultiTenancySourceFile", "could not read Multi-Tenancy File: "+err.Error())
		return err
	}

	return nil
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

	//backend.Logger.Info("client", "region", mClient.Host, "listMetrics", sortedMetadataWithMetricNames)

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
				//backend.Logger.Info("client.utils", "listMetricsPerAllRegion-region", req)

				if len(metadata) > 0 {
					//backend.Logger.Info("client.utils", "listMetricsPerAllRegion-region", sRegion)
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

	resourceTags := map[string][]string{}
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

func getComputeResourceTagsPerRegion(ctx context.Context, cClient core.ComputeClient, req core.ListInstancesRequest) (map[string][]string, map[string]map[string]struct{}) {
	backend.Logger.Debug("client.utils", "getComputeResourceTagsPerRegion", "Fetching the compute instanse tags from the oci")

	var fetchedResourceDetails []core.Instance
	var pageHeader string

	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := cClient.ListInstances(ctx, req)
		if err != nil {
			backend.Logger.Error("client.utils", "getComputeResourceTags", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	for _, item := range fetchedResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})
	}

	return fetchResourceTags(resourceTagsResponse)
}

func getVCNResourceTagsPerRegion(ctx context.Context, vClient core.VirtualNetworkClient, req core.ListVcnsRequest) (map[string][]string, map[string]map[string]struct{}) {
	backend.Logger.Debug("client.utils", "getVCNResourceTagsPerRegion", "Fetching the vcn resource tags from the oci")

	var fetchedResourceDetails []core.Vcn
	var pageHeader string

	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := vClient.ListVcns(ctx, req)
		if err != nil {
			backend.Logger.Error("client.utils", "getVCNResourceTagsPerRegion", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	for _, item := range fetchedResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})
	}

	return fetchResourceTags(resourceTagsResponse)
}

func getLBaaSResourceTagsPerRegion(ctx context.Context, lbClient loadbalancer.LoadBalancerClient, req loadbalancer.ListLoadBalancersRequest) (map[string][]string, map[string]map[string]struct{}) {
	backend.Logger.Debug("client.utils", "getVCNResourceTagsPerRegion", "Fetching the vcn resource tags from the oci")

	var fetchedResourceDetails []loadbalancer.LoadBalancer
	var pageHeader string

	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := lbClient.ListLoadBalancers(ctx, req)
		if err != nil {
			backend.Logger.Error("client.utils", "getVCNResourceTagsPerRegion", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	for _, item := range fetchedResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})
	}

	return fetchResourceTags(resourceTagsResponse)
}

func getHealthChecksTagsPerRegion(ctx context.Context, hcClient healthchecks.HealthChecksClient, req healthchecks.ListPingMonitorsRequest) (map[string][]string, map[string]map[string]struct{}) {
	backend.Logger.Debug("client.utils", "getHealthChecksTagsPerRegion", "Fetching the health check resource tags from the oci")

	var fetchedResourceDetails []healthchecks.PingMonitorSummary
	var pageHeader string

	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := hcClient.ListPingMonitors(ctx, req)
		if err != nil {
			backend.Logger.Error("client.utils", "getHealthChecksTagsPerRegion", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	for _, item := range fetchedResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})
	}

	return fetchResourceTags(resourceTagsResponse)
}
