/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

// Prepare format to decode SecureJson
func transcode(in, out interface{}) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(in)
	json.NewDecoder(buf).Decode(out)
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

// listMetricsMetadataFromAllRegion will list either metric names with namespaces or dimensions for all subscribed region
func listMetricsMetadataFromAllRegion(
	ctx context.Context,
	ci *ristretto.Cache,
	cacheKey string,
	fetchFor string,
	mClient monitoring.MonitoringClient,
	req monitoring.ListMetricsRequest,
	regions []string) map[string][]string {

	backend.Logger.Error("client.utils", "listMetricsMetadataFromAllRegion", "Data fetch start by calling list metrics API from all subscribed regions")

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

// listMetricsMetadataPerRegion will list all either dimensions or metrics either per namespaces or per resurce groups for a particular region
func listMetricsMetadataPerRegion(
	ctx context.Context,
	ci *ristretto.Cache,
	cacheKey string,
	fetchFor string,
	mClient monitoring.MonitoringClient,
	req monitoring.ListMetricsRequest) map[string][]string {

	backend.Logger.Error("client.utils", "listMetricsMetadataPerRegion", "Data fetch start by calling list metrics API for a particular regions")
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
		case constants.FETCH_FOR_LABELDIMENSION:
			for k, v := range item.Dimensions {
				// we don't need region or resource id dimensions as
				// we already filtered by region and resourceDisplayName is already there
				// in the dimensions
				// and do we really need imageId, image name will be good to have
				if k == "region" || k == "imageId" {
					continue
				}

				// to sort the final map by dimension keys
				metadata = append(metadata, k)
				metadataWithMetricNames[k] = append(metadataWithMetricNames[k], v)

				// metadataWithMetricNames[k] = []string{v}
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

// Get the tenancy Access Key
func (o *OCIDatasource) GetTenancyAccessKey(tenancyOCID string) string {

	var takey string
	tenancymode := o.settings.TenancyMode

	if tenancymode == "multitenancy" {
		takey = tenancyOCID
	} else {
		takey = SingleTenancyKey
	}

	_, ok := o.tenancyAccess[takey]
	if ok {
		backend.Logger.Error("GetTenancyAccessKey", "Valid takey", takey)
	} else {
		backend.Logger.Error("GetTenancyAccessKey", "Invalid takey", takey)
		return ""
	}

	return takey
}

// till here

func fetchResourceTags(resourceTagsResponse []models.OCIResourceTagsResponse) (map[string][]string, map[string]map[string]struct{}) {
	backend.Logger.Error("client.utils", "fetchResourceTags", "Fetching the tags from the oci call response")

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
				cacheKey := strings.Join([]string{key, tagValue}, "=")

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
			cacheKey := strings.Join([]string{k, v}, "=")

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
	backend.Logger.Error("client.utils", "collectResourceTags", "Fetching the tags from the oci call response")

	// holds key: map of values, for UI
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
				cacheKey := strings.Join([]string{tagKey, tagValue}, "=")

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
			cacheKey := strings.Join([]string{tagKey, tagValue}, "=")

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
	backend.Logger.Error("client.utils", "convertToArray", "Converting to array")

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

func extractRawDimensions(input string) []string {
	re := regexp.MustCompile(`{([^}]*)}`)
	matches := re.FindStringSubmatch(input)

	if len(matches) < 2 {
		return nil
	}

	assignments := strings.Split(matches[1], ",")
	for i, assignment := range assignments {
		assignments[i] = strings.ReplaceAll(strings.TrimSpace(assignment), " ", "")
	}

	return assignments
}

func getUniqueIdsForLabels(namespace string, dimensions map[string]string, metric string) (string, string, string, bool) {
	monitorID := ""
	var resourceID string // Declare resourceID
	var found bool

	// Define a list of keys to search in dimensions
	keys := []string{"resourceId", "ResourceId", "name"}

	// Iterate over the keys
	for _, key := range keys {
		resourceID, found = dimensions[key]
		if found {
			break
		}
	}

	// getting the extra unique id as per namespace
	if namespace == constants.OCI_NS_APM {
		resourceID = dimensions["MonitorName"]
	}

	// If resourceID is still empty, check for special conditions
	if resourceID == "" {
		// // Define a map for special conditions
		specialConditions := map[string]string{
			"node_":      "host",
			"container_": "container",
			"kube_":      "", // We will fill this in dynamically
			// 	Add more conditions here as needed
		}

		// Check each condition
		for prefix, dimension := range specialConditions {
			if strings.HasPrefix(metric, prefix) {
				// for kube metrics use the second string after _
				if prefix == "kube_" {
					split := strings.SplitN(metric, "_", 3)
					if len(split) > 2 {
						dimension = split[1]
						// job does exception
						if dimension == "job" {
							dimension = "job_name"
						}
					}
				}
				resourceID = dimensions[dimension]
				break
			}
		}

		// If no special condition was met, default to the first value in dimensions
		if resourceID == "" {
			for _, v := range dimensions {
				resourceID = v
				break
			}
		}
	}

	// some data give ocid in all caps
	resourceID = strings.ToLower(resourceID)

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

func addSelectedValuesLabels(existingLabels map[string]string, selectedValuePairs []string) map[string]string {
	if existingLabels == nil {
		existingLabels = map[string]string{}
	}

	for _, valuePair := range selectedValuePairs {
		kv := strings.Split(valuePair, "=")

		existingLabels[strings.ToLower(kv[0])] = strings.TrimPrefix(strings.TrimSuffix(kv[1], "\""), "\"")
	}

	return existingLabels
}

/*
Function generates a custom metric label for the identified metric based on the
legend format provided by the user where any known placeholders within the format
will be replaced with the appropriate value.

The currently supported legend format placeholders are:
  - {metric} - Will be replaced by the metric name
  - {dimension} - Will be replaced by the value of the specified dimension

Any placeholders (or other text) in the legend format that do not line up with one
of these placeholders will be unchanged. Note that placeholder labels are treated
as case sensitive.
*/
func (o *OCIDatasource) generateCustomMetricLabel(legendFormat string, metricName string,
	dimensions map[string][]string, UniqueDataID string) string {
	o.logger.Debug("generateCustomMetricLabel ", "legendFormat", legendFormat)
	o.logger.Debug("generateCustomMetricLabel ", "metricName", metricName)

	metricLabel := legendFormat
	// Define a pattern where we are looking for a left curly brace followed by one or
	// more characters that are not the right curly brace (or whitespace) followed
	// finally by a right curly brace. The inclusion of the <label> portion of the
	// pattern is to allow the logic to extract the label text from the placeholder.
	rePlaceholderLabel, err := regexp.Compile(`\{\{\s*(?P<label>[^} ]+)\s*\}\}`)

	if err != nil {
		o.logger.Error("Compilation of legend format placeholders regex failed")
		return metricLabel
	}

	for _, placeholderStr := range rePlaceholderLabel.FindAllString(metricLabel, -1) {
		if rePlaceholderLabel.Match([]byte(placeholderStr)) == true {
			matches := rePlaceholderLabel.FindStringSubmatch(placeholderStr)
			labelIndex := rePlaceholderLabel.SubexpIndex("label")

			placeholderLabel := matches[labelIndex]
			re := regexp.MustCompile(placeholderStr)

			// If this placeholder is the {metric} placeholder then replace the
			// placeholder string with the metric name
			if placeholderLabel == "metric" {
				metricLabel = re.ReplaceAllString(metricLabel, metricName)
			} else {

				var resourceValues []string
				var ok bool

				// Check whether there is a resourceID dimension for the metric.
				// That will be the aggregator for labeling process.
				// If not found then labeling will not be possible
				resourceValues, ok = dimensions["resourceId"]
				if !ok {
					resourceValues, ok = dimensions["resourceID"]
				}
				if !ok {
					resourceValues, ok = dimensions[placeholderLabel]
				}
				if !ok {
					return ""
				}

				// Check whether there is a dimension name for the metric that matches
				// the placeholder label. If there is then replace the placeholder with
				// the value of the dimension
				keyValues, ok := dimensions[placeholderLabel]
				if !ok {
					return ""
				}

				for i, rv := range resourceValues {
					// Check whether UniqueDataID matches any resourceID
					if rv != UniqueDataID && strings.ToLower(rv) != UniqueDataID {
						continue
					}
					sublabel := keyValues[i]
					o.logger.Debug("s_sublabel", "s_sublabel", sublabel)
					o.logger.Debug("metricLabel before", "metricLabel", metricLabel)
					metricLabel = re.ReplaceAllString(metricLabel, sublabel)
					o.logger.Debug("metricLabel after", "metricLabel", metricLabel)
				}

			}
		}
	}
	o.logger.Debug("Generated metric label", "legendFormat", legendFormat,
		"metricName", metricName, "metricLabel", metricLabel)
	return metricLabel
}
