package plugin

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
	"github.com/pkg/errors"
)

type metricDataBank struct {
	dataPoints     []monitoring.MetricData
	resourceLabels map[string]map[string]string
}

// TestConnectivity checks the OCI data source test request in Grafana's Datasource configuration UI.
//
// This function performs a connectivity test to the Oracle Cloud Infrastructure (OCI) Monitoring service.
// It verifies the configured credentials and permissions by attempting to list metrics at both the tenancy
// and compartment levels.
//
// The function iterates through each configured tenancy access key. For each key, it performs the following steps:
// 1. Fetches the tenancy OCID using the `FetchTenancyOCID` method.
// 2. Retrieves the configured region using `o.tenancyAccess[key].config.Region()`.
// 3. Attempts to list metrics at the tenancy level.
// 4. If listing metrics at the tenancy level fails, it attempts to list metrics at each compartment level.
// 5. The function checks for various error conditions and returns appropriate error messages.
//
// Parameters:
//   - ctx: The context.Context for the request.
//
// Returns:
//   - error: An error if any of the tests fail, or nil if the connectivity is successful.
func (o *OCIDatasource) TestConnectivity(ctx context.Context) error {
	// Log the start of the test
	backend.Logger.Error("client", "TestConnectivity", "testing the OCI connectivity")

	// var reg common.Region
	var testResult bool

	// Check if the tenancy access configurations are empty
	if len(o.tenancyAccess) == 0 {
		return fmt.Errorf("TestConnectivity failed: cannot read o.tenancyAccess")
	}

	// Iterate over the tenancy access configurations
	for key := range o.tenancyAccess {
		testResult = false

		// Fetch the tenancy OCID
		tenancyocid, tenancyErr := o.FetchTenancyOCID(key)
		if tenancyErr != nil {
			return errors.Wrap(tenancyErr, "error fetching TenancyOCID")
		}

		// Get the region from the tenancy access configuration
		regio, regErr := o.tenancyAccess[key].config.Region()
		if regErr != nil {
			return errors.Wrap(regErr, "error fetching Region")
		} else {
			backend.Logger.Debug("TestConnectivity", "ConfigKey", key, "Region", regio)
		}

		// Test the tenancy OCID
		backend.Logger.Error("TestConnectivity", "ConfigKey", key, "Testing Tenancy OCID", tenancyocid)
		listMetrics := monitoring.ListMetricsRequest{
			CompartmentId: &tenancyocid,
			Limit:         common.Int(25),
		}

		var status int
		res, err := o.tenancyAccess[key].monitoringClient.ListMetrics(ctx, listMetrics)
		if res.RawResponse == nil || res.RawResponse.ContentLength == 0 {
			backend.Logger.Error("TestConnectivity", "Config Key", key, "error", err)
			return fmt.Errorf("TestConnectivity failed: result is empty %v: %v", key, err)
		} else {
			status = res.RawResponse.StatusCode
		}

		// Handle errors
		if err != nil {
			if res.RawResponse.StatusCode == 401 {
				backend.Logger.Error("TestConnectivity", "Config Key", key, "error", err)
				return fmt.Errorf("TestConnectivity failed: error in profile %v: %v", key, err)
			} else {
				backend.Logger.Error("TestConnectivity", "Config Key", key, "SKIPPED", err)
			}
		}

		// Check the status code
		if status >= 200 && status < 300 {
			backend.Logger.Error("TestConnectivity", "Config Key", key, "OK", status)
		} else {
			backend.Logger.Error("TestConnectivity", "Config Key", key, "SKIPPED", fmt.Sprintf("listMetrics on Tenancy %s did not work, testing compartments", tenancyocid))

			// Get the compartments
			comparts := o.GetCompartments(ctx, tenancyocid, true)
			if comparts == nil {
				backend.Logger.Error("TestConnectivity", "Config Key", key, "error", "could not read compartments")
				return fmt.Errorf("TestConnectivity failed: cannot read Compartments in profile %v", key)
			}

			// Test each compartment
			for _, v := range comparts {
				tocid := v.OCID
				backend.Logger.Error("TestConnectivity", "Config Key", key, "Testing", tocid)
				listMetrics := monitoring.ListMetricsRequest{
					CompartmentId: common.String(tocid),
					Limit:         common.Int(25),
				}

				res, err := o.tenancyAccess[key].monitoringClient.ListMetrics(ctx, listMetrics)
				if err != nil {
					backend.Logger.Error("TestConnectivity", "Config Key", key, "SKIPPED", err)
				}
				status := res.RawResponse.StatusCode
				if status >= 200 && status < 300 {
					backend.Logger.Error("TestConnectivity", "Config Key", key, "OK", status)
					testResult = true
					break
				} else {
					backend.Logger.Error("TestConnectivity", "Config Key", key, "SKIPPED", status)
				}
			}
			if testResult {
				continue
			} else {
				backend.Logger.Error("TestConnectivity", "Config Key", key, "FAILED", "listMetrics failed in each compartment")
				return fmt.Errorf("listMetrics failed in each Compartments in profile %v", key)
			}
		}
	}
	return nil
}

/*
FetchTenancyOCID retrieves the tenancy OCID based on the provided tenancy access key (takey).

This function handles different tenancy modes (single vs. multi-tenancy) and environments (local vs. OCI Instance).
It fetches the tenancy OCID from the appropriate configuration provider.

Parameters:
  - takey: The tenancy access key.

Returns:
  - string: The tenancy OCID.
  - error: An error if the tenancy OCID cannot be fetched.
*/
func (o *OCIDatasource) FetchTenancyOCID(takey string) (string, error) {
	tenv := o.settings.Environment
	tenancymode := o.settings.TenancyMode
	xtenancy := o.settings.Xtenancy_0
	var tenancyocid string
	var tenancyErr error

	if tenancymode == "multitenancy" && tenv == "OCI Instance" {
		return "", errors.New("Multitenancy mode using instance principals is not implemented yet.")
	}

	if tenancymode == "multitenancy" {
		if len(takey) <= 0 || takey == NoTenancy {
			o.logger.Error("Unable to get Multi-tenancy OCID")
			return "", errors.Wrap(tenancyErr, "error fetching TenancyOCID")
		} else {
			res := strings.Split(takey, "/")
			tenancyocid = res[1]
		}
	} else {
		if xtenancy != "" && tenv == "OCI Instance" {
			o.logger.Debug("Cross Tenancy Instance Principal detected")
			tocid, _ := o.tenancyAccess[takey].config.TenancyOCID()
			o.logger.Debug("Source Tenancy OCID: " + tocid)
			o.logger.Debug("Target Tenancy OCID: " + o.settings.Xtenancy_0)
			tenancyocid = xtenancy
		} else {
			tenancyocid, tenancyErr = o.tenancyAccess[takey].config.TenancyOCID()
			if tenancyErr != nil {
				return "", errors.Wrap(tenancyErr, "error fetching TenancyOCID")
			}
		}
	}
	return tenancyocid, nil
}

/*
GetTenancies function

Generates an array containing OCI tenancy list in the following format:
<Label/TenancyOCID>

This function retrieves the list of tenancies available in the OCI environment.

Parameters:
  - ctx: The context.Context for the request.

Returns:
  - []models.OCIResource: A slice of OCIResource containing tenancy information.
*/
func (o *OCIDatasource) GetTenancies(ctx context.Context) []models.OCIResource {
	backend.Logger.Error("client", "GetTenancies", "fetching the tenancies")

	tenancyList := []models.OCIResource{}
	for key := range o.tenancyAccess {
		// frame.AppendRow(*(common.String(key)))

		tenancyList = append(tenancyList, models.OCIResource{
			Name: *(common.String(key)),
			OCID: *(common.String(key)),
		})
	}

	return tenancyList
}

// GetSubscribedRegions Returns the subscribed regions by the mentioned tenancy
// API Operation: ListRegionSubscriptions
// Permission Required: TENANCY_INSPECT
// Links:
// https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/iampolicyreference.htm
// https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/managingregions.htm
// https://docs.oracle.com/en-us/iaas/api/#/en/identity/20160918/RegionSubscription/ListRegionSubscriptions
//
// This function retrieves the list of regions subscribed to by a specific tenancy in Oracle Cloud Infrastructure.
// It queries the Identity service to obtain the list of subscribed regions.
//
// Parameters:
//   - ctx: The context.Context for the request.
//   - tenancyOCID: The OCID of the tenancy for which to list subscribed regions.
//
// Returns:
//   - []string: A slice of strings, where each string represents a subscribed region.
//     Returns nil if any error occurred during the process.
func (o *OCIDatasource) GetSubscribedRegions(ctx context.Context, tenancyOCID string) []string {
	backend.Logger.Error("client", "GetSubscribedRegions", "fetching the subscribed region for tenancy: "+tenancyOCID)

	var subscribedRegions []string
	takey := o.GetTenancyAccessKey(tenancyOCID)

	if len(takey) == 0 {
		backend.Logger.Warn("client", "GetSubscribedRegions", "invalid takey")
		return nil
	}

	tenancyocid, tenancyErr := o.FetchTenancyOCID(takey)
	if tenancyErr != nil {
		backend.Logger.Warn("client", "GetSubscribedRegions", tenancyErr)
		return nil
	}

	backend.Logger.Error("client", "GetSubscribedRegionstakey", "fetching the subscribed region for tenancy OCID: "+*common.String(tenancyocid))

	req := identity.ListRegionSubscriptionsRequest{TenancyId: common.String(tenancyocid)}

	resp, err := o.tenancyAccess[takey].identityClient.ListRegionSubscriptions(ctx, req)
	if err != nil {
		backend.Logger.Warn("client", "GetSubscribedRegions", err)
		return nil
	}

	// if err != nil {
	// 	backend.Logger.Warn("client", "GetSubscribedRegions", err)
	// 	subscribedRegions = append(subscribedRegions, o.tenancyAccess[takey].region)
	// 	return subscribedRegions
	// }
	if resp.RawResponse.StatusCode != 200 {
		backend.Logger.Warn("client", "GetSubscribedRegions", "Could not fetch subscribed regions. Please check IAM policy.")
		return subscribedRegions
	}

	for _, item := range resp.Items {
		if item.Status == identity.RegionSubscriptionStatusReady {
			backend.Logger.Error("client", "GetSubscribedRegionstakey", "fetching the subscribed region for regioname: "+*item.RegionName)
			subscribedRegions = append(subscribedRegions, *item.RegionName)
		}
	}

	if len(subscribedRegions) > 1 {
		subscribedRegions = append(subscribedRegions, constants.ALL_REGION)
	}
	/* Sort regions list */
	sort.Strings(subscribedRegions)
	return subscribedRegions
}

// GetCompartments Returns all the sub compartments under the tenancy
// API Operation: ListCompartments
// Permission Required: COMPARTMENT_INSPECT
// Links:
// https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/iampolicyreference.htm
// https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/managingcompartments.htm
// https://docs.oracle.com/en-us/iaas/api/#/en/identity/20160918/Compartment/ListCompartments
//
// Retrieves a list of compartments within a specified tenancy in Oracle Cloud Infrastructure.
// This function interacts with the OCI Identity service to fetch compartments.
//
// Parameters:
//   - ctx: The context.Context for the request.
//   - tenancyOCID: The OCID of the tenancy for which to list compartments.
//   - includeAccessibleOnly: An optional boolean flag. If true, only accessible compartments are included.
//     If omitted or false, all compartments are included.
//
// Returns:
//   - []models.OCIResource: A slice of OCIResource, where each element represents a compartment with its
//     name and OCID. Returns nil if there is an error during the process.
func (o *OCIDatasource) GetCompartments(ctx context.Context, tenancyOCID string, includeAccessibleOnly ...bool) []models.OCIResource {
	backend.Logger.Error("client", "GetCompartments", "fetching the sub-compartments for tenancy: "+tenancyOCID)

	takey := o.GetTenancyAccessKey(tenancyOCID)

	tenancyocid, tenancyErr := o.FetchTenancyOCID(takey)
	if tenancyErr != nil {
		backend.Logger.Warn("client", "GetSubscribedRegions", tenancyErr)
		return nil
	}

	// fetching from cache, if present
	cacheKey := strings.Join([]string{tenancyocid, "cs"}, "-")
	if cachedCompartments, found := o.cache.Get(cacheKey); found {
		backend.Logger.Warn("client", "GetCompartments", "getting the data from cache")
		return cachedCompartments.([]models.OCIResource)
	}

	req := identity.GetTenancyRequest{TenancyId: common.String(tenancyocid)}

	// Send the request using the service client
	resp, err := o.tenancyAccess[takey].identityClient.GetTenancy(context.Background(), req)
	if err != nil {
		backend.Logger.Error("client", "GetCompartments", "error in GetTenancy")
		return nil
	}

	var effectiveScope identity.ListCompartmentsAccessLevelEnum

	if len(includeAccessibleOnly) == 1 && includeAccessibleOnly[0] {
		effectiveScope = identity.ListCompartmentsAccessLevelAccessible
		backend.Logger.Error("client", "GetCompartments", "using ListCompartmentsAccessLevelAccessible")
	} else {
		effectiveScope = identity.ListCompartmentsAccessLevelAny
	}

	compartments := map[string]string{}

	// calling the api if not present in cache
	compartmentList := []models.OCIResource{}
	var fetchedCompartments []identity.Compartment
	var pageHeader string

	for {
		res, err := o.tenancyAccess[takey].identityClient.ListCompartments(ctx,
			identity.ListCompartmentsRequest{
				CompartmentId:          common.String(tenancyocid),
				Page:                   &pageHeader,
				AccessLevel:            effectiveScope,
				LifecycleState:         identity.CompartmentLifecycleStateActive,
				CompartmentIdInSubtree: common.Bool(true),
			})

		if err != nil {
			backend.Logger.Warn("client", "GetCompartments", err)
			break
		}

		fetchedCompartments = append(fetchedCompartments, res.Items...)

		if len(res.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *res.OpcNextPage
		} else {
			break
		}
	}

	compartments[tenancyocid] = *resp.Name //tenancy name

	// storing compartment ocid and name
	for _, item := range fetchedCompartments {
		compartments[*item.Id] = *item.Name
	}

	// checking if parent compartment is there or not, and update name accordingly
	for _, item := range fetchedCompartments {
		compartmentName := *item.Name
		compartmentOCID := *item.Id
		parentCompartmentOCID := *item.CompartmentId

		if pcn, found := compartments[parentCompartmentOCID]; found {
			compartmentName = pcn + " > " + compartmentName
		}

		compartmentList = append(compartmentList, models.OCIResource{
			Name: compartmentName,
			OCID: compartmentOCID,
		})
	}

	compartmentList = append(compartmentList, models.OCIResource{
		Name: *resp.Name,
		OCID: tenancyocid,
	})

	// if len(compartmentList) > 1 {
	// 	compartmentList = append(compartmentList, models.OCIResource{
	// 		Name: constants.ALL_COMPARTMENT,
	// 		OCID: "",
	// 	})
	// }

	// sorting based on compartment name
	sort.SliceStable(compartmentList, func(i, j int) bool {
		return compartmentList[i].Name < compartmentList[j].Name
	})

	// saving in the cache
	o.cache.SetWithTTL(cacheKey, compartmentList, 1, 15*time.Minute)
	o.cache.Wait()

	return compartmentList
}

// GetNamespaceWithMetricNames retrieves a list of namespaces along with their associated metric names within a specified compartment of an OCI tenancy.
//
// This function interacts with the OCI Monitoring service to fetch the namespaces and their respective metric names.
// It supports caching to improve performance and reduces the number of API calls to OCI. It also supports fetching data for all subscribed regions.
//
// Parameters:
//   - ctx: The context.Context for the request, used for cancellation and request-scoped values.
//   - tenancyOCID: The OCID of the tenancy in which to search for namespaces and metrics.
//   - compartmentOCID: The OCID of the compartment in which to search. If empty, the search spans the entire tenancy.
//   - region: The OCI region to search in. If constants.ALL_REGION is specified, data from all subscribed regions is fetched.
//
// Returns:
//   - []models.OCIMetricNamesWithNamespace: A slice of OCIMetricNamesWithNamespace, where each element contains a namespace and its associated metric names.
//   - An empty slice if no namespaces or metrics are found, or if an error occurs.
//
// API Operation:
//   - ListMetrics: https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/Metric/ListMetrics
//
// Permissions Required:
//   - METRIC_INSPECT: Required to list metrics and namespaces.
//
// Caching:
//   - The results are cached to reduce API calls. The cache key is generated using the tenancy OCID, compartment OCID, region, and the string "nss".
//   - Cached data has a Time To Live (TTL) of 5 minutes.
//
// Error Handling:
//   - Logs errors encountered during the process.
//   - Returns an empty slice if errors occur or no data is found.
func (o *OCIDatasource) GetNamespaceWithMetricNames(
	ctx context.Context,
	tenancyOCID string,
	compartmentOCID string,
	region string) []models.OCIMetricNamesWithNamespace {
	backend.Logger.Error("client", "GetNamespaceWithMetricNames", "fetching the metric names along with namespaces under compartment: "+compartmentOCID)

	takey := o.GetTenancyAccessKey(tenancyOCID)
	// fetching from cache, if present
	cacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, "nss"}, "-")
	if cachedMetricNamesWithNamespaces, found := o.cache.Get(cacheKey); found {
		// This check avoids the type assertion and potential panic
		if _, ok := cachedMetricNamesWithNamespaces.([]models.OCIMetricNamesWithNamespace); ok {
			backend.Logger.Warn("client", "GetNamespaceWithMetricNames", "getting the data from cache")
			return cachedMetricNamesWithNamespaces.([]models.OCIMetricNamesWithNamespace)
		} else {
			backend.Logger.Warn("client.utils", "GetNamespaceWithMetricNames", "cannot use cached data -> "+cacheKey)
		}
	}

	// calling the api if not present in cache
	var namespaceWithMetricNames map[string][]string
	namespaceWithMetricNamesList := []models.OCIMetricNamesWithNamespace{}

	monitoringRequest := monitoring.ListMetricsRequest{
		CompartmentId:          common.String(compartmentOCID),
		CompartmentIdInSubtree: common.Bool(false),
		ListMetricsDetails: monitoring.ListMetricsDetails{
			GroupBy:   []string{"namespace", "name"},
			SortBy:    monitoring.ListMetricsDetailsSortByNamespace,
			SortOrder: monitoring.ListMetricsDetailsSortOrderAsc,
		},
	}

	// when search is wide along the tenancy
	if len(compartmentOCID) == 0 {
		monitoringRequest.CompartmentId = common.String(tenancyOCID)
		monitoringRequest.CompartmentIdInSubtree = common.Bool(true)
	}

	// when user wants to fetch everything for all subscribed regions
	if region == constants.ALL_REGION {
		namespaceWithMetricNames = listMetricsMetadataFromAllRegion(
			ctx,
			o.cache,
			cacheKey,
			constants.FETCH_FOR_NAMESPACE,
			o.tenancyAccess[takey].monitoringClient,
			monitoringRequest,
			o.GetSubscribedRegions(ctx, tenancyOCID),
		)
	} else {
		namespaceWithMetricNames = listMetricsMetadataPerRegion(
			ctx,
			o.cache,
			cacheKey,
			constants.FETCH_FOR_NAMESPACE,
			o.tenancyAccess[takey].monitoringClient,
			monitoringRequest,
		)
	}

	// preparing for frontend
	for k, v := range namespaceWithMetricNames {
		namespaceWithMetricNamesList = append(namespaceWithMetricNamesList, models.OCIMetricNamesWithNamespace{
			Namespace:   k,
			MetricNames: v,
		})
	}

	// sort namespace
	sort.Slice(namespaceWithMetricNamesList, func(i, j int) bool {
		return namespaceWithMetricNamesList[i].Namespace < namespaceWithMetricNamesList[j].Namespace
	})

	// saving into the cache
	o.cache.SetWithTTL(cacheKey, namespaceWithMetricNamesList, 1, 5*time.Minute)
	o.cache.Wait()

	return namespaceWithMetricNamesList
}

// GetMetricDataPoints retrieves metric data points from the OCI Monitoring service based on the provided parameters.
//
// This function queries the OCI Monitoring service to retrieve aggregated metric data points. It supports various filters
// such as tenancy, compartment, namespace, query text, time range, resource group, dimensions, and tags. The function also
// handles fetching data across multiple regions in parallel and performs necessary data adjustments for accurate representation.
//
// Parameters:
//   - ctx: The context.Context for the request.
//   - requestParams: A models.MetricsDataRequest struct containing all the necessary parameters for the query.
//   - tenancyOCID: The OCID of the tenancy for which the metrics data is requested.
//
// Returns:
//   - []time.Time: A slice of time.Time representing the timestamps for the retrieved data points.
//   - []models.OCIMetricDataPoints: A slice of OCIMetricDataPoints, each containing the data points, labels, and other metadata for a metric.
//   - error: An error if any operation fails during the process.
//
// API Operation:
//   - SummarizeMetricsData: https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/MetricData/SummarizeMetricsData
//
// Permissions Required:
//   - METRIC_INSPECT: Required to inspect metrics.
//   - METRIC_READ: Required to read metric data.
//
// Data Handling:
//   - Handles fetching data for all regions in parallel when specified.
//   - Adjusts data when different resource datapoints have a mismatched number of values.
//   - Supports filtering by resource group, dimensions, and tags.
//   - Adds labels based on selected dimensions and tags.
//   - Sorts the time slice for proper representation in Grafana.
//
// Error Handling:
//   - Returns an error if an invalid 'takey' (tenancy access key) is detected.
//   - Returns any errors encountered during API calls.
//   - Logs errors encountered during the data retrieval process.
func (o *OCIDatasource) GetMetricDataPoints(ctx context.Context, requestParams models.MetricsDataRequest, tenancyOCID string) ([]time.Time, []models.OCIMetricDataPoints, error) {
	backend.Logger.Error("client", "GetMetricDataPoints", "fetching the metrics datapoints under compartment '"+requestParams.CompartmentOCID+"' for query '"+requestParams.QueryText+"'")

	times := []time.Time{}
	dataValuesWithTime := map[common.SDKTime][]float64{}
	dataPointsWithResourceSerialNo := map[int]models.OCIMetricDataPoints{}
	dataPoints := []models.OCIMetricDataPoints{}
	resourceIDsPerTag := map[string]map[string]struct{}{}
	var takey string
	selectedTags := requestParams.TagsValues
	selectedDimensions := requestParams.DimensionValues

	if tenancyOCID == "select tenancy" && o.settings.TenancyMode == "single" {
		takey = o.GetTenancyAccessKey("DEFAULT/")
	} else {
		takey = o.GetTenancyAccessKey(tenancyOCID)
	}

	if len(takey) == 0 {
		backend.Logger.Warn("client", "GetMetricDataPoints", "invalid takey")
		return nil, nil, errors.New("Datasource not configured (invalid takey)")
	}

	metricsDataRequest := monitoring.SummarizeMetricsDataRequest{
		CompartmentId:          common.String(requestParams.CompartmentOCID),
		CompartmentIdInSubtree: common.Bool(false),
		SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
			Namespace: common.String(requestParams.Namespace),
			Query:     common.String(requestParams.QueryText),
			StartTime: &common.SDKTime{Time: requestParams.StartTime},
			EndTime:   &common.SDKTime{Time: requestParams.EndTime},
		},
	}

	// to search for all compartments
	if len(requestParams.CompartmentOCID) == 0 {
		metricsDataRequest.CompartmentId = common.String(requestParams.TenancyOCID)
		metricsDataRequest.CompartmentIdInSubtree = common.Bool(true)
	}

	// adding the resource group when provided
	if len(requestParams.ResourceGroup) != 0 {
		if requestParams.ResourceGroup != constants.DEFAULT_RESOURCE_PLACEHOLDER && requestParams.ResourceGroup != constants.DEFAULT_RESOURCE_PLACEHOLDER_LEGACY && requestParams.ResourceGroup != constants.DEFAULT_RESOURCE_GROUP {
			metricsDataRequest.SummarizeMetricsDataDetails.ResourceGroup = &requestParams.ResourceGroup
		}
	}

	var allRegionsMetricsDataPoint sync.Map
	subscribedRegions := []string{}

	if requestParams.Region == constants.ALL_REGION {
		subscribedRegions = append(subscribedRegions, o.GetSubscribedRegions(ctx, requestParams.TenancyOCID)...)
	} else {
		if requestParams.Region != "" {
			subscribedRegions = append(subscribedRegions, requestParams.Region)
		}
	}

	// fetching the metrics data for specified region in parallel
	var wg sync.WaitGroup
	errCh := make(chan error)
	for _, subscribedRegion := range subscribedRegions {
		if subscribedRegion != constants.ALL_REGION {
			wg.Add(1)
			go func(mc monitoring.MonitoringClient, sRegion string, errCh chan error) {
				defer wg.Done()
				resp, err := mc.SummarizeMetricsData(ctx, metricsDataRequest)
				if err != nil {
					backend.Logger.Error("client", "GetMetricDataPoints", err)
					errCh <- err
				}

				if len(resp.Items) > 0 {
					// fetching the resource labels
					var rl map[string]map[string]string

					// Tags will be used in future releases
					// cachedResourceLabels := o.fetchFromCache(
					// 	ctx,
					// 	requestParams.TenancyOCID,
					// 	requestParams.CompartmentOCID,
					// 	requestParams.CompartmentName,
					// 	sRegion,
					// 	requestParams.Namespace,
					// 	"resource_labels",
					// )

					// rl = cachedResourceLabels.(map[string]map[string]string)

					// storing the data to calculate later
					allRegionsMetricsDataPoint.Store(sRegion, metricDataBank{
						dataPoints:     resp.Items,
						resourceLabels: rl,
					})
				}
				errCh <- nil
			}(o.tenancyAccess[takey].monitoringClient, subscribedRegion, errCh)
			// Receive on the error channel
			err := <-errCh
			if err != nil {
				backend.Logger.Error("client", "GetMetricDataPoints", err)
				return nil, nil, err
			}
		}
	}
	wg.Wait()

	resourcesFetched := 0

	allRegionsMetricsDataPoint.Range(func(key, value interface{}) bool {
		regionInUse := key.(string)

		backend.Logger.Debug("client", "GetMetricDataPoints", "Metric datapoints got for region-"+regionInUse)

		// Tags will be used in future releases
		// get the selected tags
		// if len(selectedTags) != 0 {
		// 	cachedResourceNamesPerTag := o.fetchFromCache(
		// 		ctx,
		// 		requestParams.TenancyOCID,
		// 		requestParams.CompartmentOCID,
		// 		requestParams.CompartmentName,
		// 		regionInUse,
		// 		requestParams.Namespace,
		// 		constants.CACHE_KEY_RESOURCE_IDS_PER_TAG,
		// 	)

		// 	resourceIDsPerTag = cachedResourceNamesPerTag.(map[string]map[string]struct{})
		// }

		metricData := value.(metricDataBank)

		for _, metricDataItem := range metricData.dataPoints {
			found := false

			uniqueDataID, dimensionKey, resourceDisplayName, extraUniqueID, rIDPresent := getUniqueIdsForLabels(requestParams.Namespace, metricDataItem.Dimensions, requestParams.QueryText)

			if rIDPresent {
				for _, selectedTag := range selectedTags {
					if _, ok := resourceIDsPerTag[selectedTag][uniqueDataID]; ok {
						found = true
						break
					}
				}

				if len(selectedTags) != 0 && !found {
					continue
				}
			}

			metricDatapoints := metricDataItem.AggregatedDatapoints

			// sorting the data by increasing time
			sort.SliceStable(metricDatapoints, func(i, j int) bool {
				return metricDatapoints[i].Timestamp.Time.Before(metricDatapoints[j].Timestamp.Time)
			})

			// sometimes 2 different resource datapoint have mismatched no of values
			// to make it equal fill the extra point with previous value
			resourcesFetched += 1
			previousValue := 0.0
			for _, eachMetricDataPoint := range metricDatapoints {
				t := *eachMetricDataPoint.Timestamp
				v := *eachMetricDataPoint.Value

				if _, ok := dataValuesWithTime[t]; ok {
					dataValuesWithTime[t] = append(dataValuesWithTime[t], v)
					previousValue = v
				} else {
					if resourcesFetched == 1 {
						dataValuesWithTime[t] = []float64{v}
						previousValue = v
						continue
					}

					// adjustment for previous non-existance values with the immediate previous value
					// when the time comes in later data points
					dataValuesWithTime[t] = []float64{previousValue}
					for i := 2; i < resourcesFetched; i++ {
						dataValuesWithTime[t] = append(dataValuesWithTime[t], previousValue)
					}
					dataValuesWithTime[t] = append(dataValuesWithTime[t], v)
					previousValue = v
				}
			}

			// for base tenancy
			splits := strings.Split(tenancyOCID, "/")
			tenancyName := splits[0]

			// to get the resource labels
			labelKey := uniqueDataID + extraUniqueID
			if strings.Contains(resourceDisplayName, "ocid") {
				resourceDisplayName = metricData.resourceLabels[labelKey]["resource_name"]
			}

			var labelsToAdd map[string]string
			if requestParams.RawQuery {
				// adding the selected dimensions as labels if dropdowns are selected
				labelsToAdd = addSelectedValuesLabels(metricData.resourceLabels[labelKey], selectedDimensions)
			} else {
				// adding the all returned dimensions as labels if raw query is selected are selected
				labelsToAdd = metricDataItem.Dimensions
			}

			// adding the selected tags as labels
			labelsToAdd = addSelectedValuesLabels(labelsToAdd, selectedTags)

			// preparing the metric data to display
			dataPointsWithResourceSerialNo[resourcesFetched-1] = models.OCIMetricDataPoints{
				TenancyName:  tenancyName,
				Region:       regionInUse,
				MetricName:   *metricDataItem.Name,
				ResourceName: resourceDisplayName,
				UniqueDataID: uniqueDataID,
				DimensionKey: dimensionKey,
				Labels:       labelsToAdd,
			}
		}

		return true
	})

	timesToFetch := []common.SDKTime{}
	// adjustment for later non-existance values with last value
	for t, dvs := range dataValuesWithTime {
		times = append(times, t.Time)
		timesToFetch = append(timesToFetch, t)

		if len(dvs) == resourcesFetched {
			continue
		}

		lastValue := dataValuesWithTime[t][len(dataValuesWithTime[t])-1]
		for i := 0; i < resourcesFetched-len(dvs); i++ {
			dataValuesWithTime[t] = append(dataValuesWithTime[t], lastValue)
		}
	}

	// sorting the time slice, for grafana
	sort.SliceStable(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	// sorting the time slice, for internal fetch
	sort.SliceStable(timesToFetch, func(i, j int) bool {
		return timesToFetch[i].Time.Before(timesToFetch[j].Time)
	})

	dataValuesWithResourceSerialNo := map[int][]float64{}
	// final preparation
	for _, t := range timesToFetch {
		dvIndex := 0
		for i := 0; i < resourcesFetched; i++ {
			dataValuesWithResourceSerialNo[i] = append(dataValuesWithResourceSerialNo[i], dataValuesWithTime[t][dvIndex])
			dvIndex += 1
		}
	}

	// extracting for grafana
	for i, dps := range dataValuesWithResourceSerialNo {
		dp := dataPointsWithResourceSerialNo[i]
		dp.DataPoints = dps

		dataPoints = append(dataPoints, dp)
	}

	return times, dataPoints, nil
}

// ****** WARNING This function is not implemented yet ******
// fetchFromCache retrieves data from the cache based on the provided parameters.
// If the data is not found in the cache, it fetches the tags and updates the cache.
//
// Parameters:
//
//	ctx - The context for controlling the request lifetime.
//	tenancyOCID - The OCID of the tenancy.
//	compartmentOCID - The OCID of the compartment.
//	compartmentName - The name of the compartment.
//	region - The region identifier.
//	namespace - The namespace identifier.
//	suffix - The suffix to be appended to the cache key.
//
// Returns:
//
//	An interface{} containing the cached resource.
func (o *OCIDatasource) fetchFromCache(ctx context.Context, tenancyOCID string, compartmentOCID string, compartmentName string, region string, namespace string, suffix string) interface{} {
	backend.Logger.Error("client", "fetchFromCache", "fetching from cache")

	labelCacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, namespace, suffix}, "-")
	if _, found := o.cache.Get(labelCacheKey); !found {
		o.GetTags(ctx, tenancyOCID, compartmentOCID, compartmentName, region, namespace)
	}

	cachedResource, _ := o.cache.Get(labelCacheKey)
	return cachedResource
}

// ****** WARNING This function is not implemented yet ******
// GetTags Returns all the defined as well as freeform tags attached with resources for a namespace under a compartment
// fetching the resources based on which type resources we want
// API Operation: ListInstances, ListVcns
// Permission Required:
// Links:
// https://docs.oracle.com/en-us/iaas/api/#/en/iaas/20160918/Instance/ListInstances
func (o *OCIDatasource) GetTags(
	ctx context.Context,
	tenancyOCID string,
	compartmentOCID string,
	compartmentName string,
	region string,
	namespace string) []models.OCIResourceTags {
	backend.Logger.Error("client", "GetTags", "fetching the tags for namespace '"+namespace+"'")

	resourceTagsList := []models.OCIResourceTags{}
	allResourceTags := map[string][]string{}

	// building the regions list
	subscribedRegions := []string{}
	if region == constants.ALL_REGION {
		subscribedRegions = append(subscribedRegions, o.GetSubscribedRegions(ctx, tenancyOCID)...)
	} else {
		if region != "" {
			subscribedRegions = append(subscribedRegions, region)
		}
	}

	// compartments := []models.OCIResource{}
	// if len(compartmentOCID) == 0 {
	// 	compartments = append(compartments, o.GetCompartments(ctx, tenancyOCID)...)
	// } else {
	// 	compartments = append(compartments, models.OCIResource{
	// 		Name: compartmentName,
	// 		OCID: compartmentOCID,
	// 	})
	// }

	// var ccc core.ComputeClient
	// //var vcc core.VirtualNetworkClient
	// var lbc loadbalancer.LoadBalancerClient
	// var hcc healthchecks.HealthChecksClient
	// var dbc database.DatabaseClient
	// var adc apmcontrolplane.ApmDomainClient
	// var asc apmsynthetics.ApmSyntheticClient
	var cErr error

	// switch constants.OCI_NAMESPACES[namespace] {
	// case constants.OCI_TARGET_COMPUTE, constants.OCI_TARGET_VCN:
	// 	ccc, cErr = client.GetComputeClient()
	// // case constants.OCI_TARGET_VCN:
	// // 	ccc, cErr = client.GetComputeClient()
	// // 	vcc, cErr = client.GetVCNClient()
	// case constants.OCI_TARGET_LBAAS:
	// 	lbc, cErr = client.GetLBaaSClient()
	// case constants.OCI_TARGET_HEALTHCHECK:
	// 	hcc, cErr = client.GetHealthChecksClient()
	// case constants.OCI_TARGET_DATABASE:
	// 	dbc, cErr = client.GetDatabaseClient()
	// case constants.OCI_TARGET_APM:
	// 	adc, asc, cErr = client.GetApmClients()
	// }

	var allRegionsResourceTags sync.Map
	var wg sync.WaitGroup
	for _, subscribedRegion := range subscribedRegions {
		if subscribedRegion != constants.ALL_REGION {
			wg.Add(1)
			go func(sRegion string) {
				defer wg.Done()

				rTagsCacheKey := strings.Join([]string{
					tenancyOCID,
					compartmentOCID,
					sRegion,
					namespace,
					constants.CACHE_KEY_RESOURCE_TAGS,
				}, "-")
				rIDsPerTagCacheKey := strings.Join([]string{
					tenancyOCID,
					compartmentOCID,
					sRegion,
					namespace,
					constants.CACHE_KEY_RESOURCE_IDS_PER_TAG,
				}, "-")

				// checking if the cache already exists
				if rawResourceTags, foundTags := o.cache.Get(rTagsCacheKey); foundTags {
					if _, foundNames := o.cache.Get(rIDsPerTagCacheKey); foundNames {
						resourceTags := rawResourceTags.(map[string][]string)
						allRegionsResourceTags.Store(sRegion, resourceTags)

						return
					}
				}

				// when creating client has some error
				if cErr != nil {
					return
				}

				labelCacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, sRegion, namespace, "resource_labels"}, "-")

				resourceTags := map[string][]string{}
				resourceIDsPerTag := map[string]map[string]struct{}{}
				resourceLabels := map[string]map[string]string{}

				// Tags to be implemented in future release
				// switch constants.OCI_NAMESPACES[namespace] {
				// case constants.OCI_TARGET_COMPUTE:
				// 	ccc.SetRegion(sRegion)
				// 	ocic := OCICore{
				// 		ctx:           ctx,
				// 		computeClient: ccc,
				// 	}
				// 	resourceTags, resourceIDsPerTag, resourceLabels = ocic.GetComputeResourceTagsPerRegion(compartments)
				// case constants.OCI_TARGET_VCN:
				// 	//vcc.SetRegion(sRegion)
				// 	ccc.SetRegion(sRegion)
				// 	ocic := OCICore{
				// 		ctx:           ctx,
				// 		computeClient: ccc,
				// 	}
				// 	resourceTags, resourceIDsPerTag, resourceLabels = ocic.GetVNicResourceTagsPerRegion(compartments)
				// case constants.OCI_TARGET_LBAAS:
				// 	lbc.SetRegion(sRegion)
				// 	ocilb := OCILoadBalancer{
				// 		ctx:    ctx,
				// 		client: lbc,
				// 	}
				// 	resourceTags, resourceIDsPerTag, resourceLabels = ocilb.GetLBaaSResourceTagsPerRegion(compartments)
				// case constants.OCI_TARGET_HEALTHCHECK:
				// 	hcc.SetRegion(sRegion)
				// 	ocihc := OCIHealthChecks{
				// 		ctx:               ctx,
				// 		healthCheckClient: hcc,
				// 	}
				// 	resourceTags, resourceIDsPerTag, resourceLabels = ocihc.GetHealthChecksTagsPerRegion(compartments)
				// case constants.OCI_TARGET_DATABASE:
				// 	dbc.SetRegion(sRegion)
				// 	db := OCIDatabase{
				// 		ctx:    ctx,
				// 		client: dbc,
				// 	}

				// 	switch namespace {
				// 	case constants.OCI_NS_DB_ORACLE:
				// 		resourceTags, resourceIDsPerTag, resourceLabels = db.GetOracleDatabaseTagsPerRegion(compartments)
				// 	case constants.OCI_NS_DB_EXTERNAL:
				// 		resourceTags, resourceIDsPerTag, resourceLabels = db.GetExternalPluggableDatabaseTagsPerRegion(compartments)
				// 	case constants.OCI_NS_DB_AUTONOMOUS:
				// 		resourceTags, resourceIDsPerTag, resourceLabels = db.GetAutonomousDatabaseTagsPerRegion(compartments)
				// 	}
				// case constants.OCI_TARGET_APM:
				// 	adc.SetRegion(sRegion)
				// 	asc.SetRegion(sRegion)
				// 	apm := OCIApm{
				// 		ctx:             ctx,
				// 		domainClient:    adc,
				// 		syntheticClient: asc,
				// 	}
				// 	resourceTags, resourceIDsPerTag, resourceLabels = apm.GetApmTagsPerRegion(compartments)
				// }

				// storing the labels in cache to use along with metric data
				o.cache.SetWithTTL(labelCacheKey, resourceLabels, 1, 15*time.Minute)
				// saving in cache - previous was 30
				o.cache.SetWithTTL(rTagsCacheKey, resourceTags, 1, 15*time.Minute)
				o.cache.SetWithTTL(rIDsPerTagCacheKey, resourceIDsPerTag, 1, 15*time.Minute)
				o.cache.Wait()

				// to store all resource tags for all region
				allRegionsResourceTags.Store(sRegion, resourceTags)
			}(subscribedRegion)
		}
	}
	wg.Wait()

	// // clearing up
	// ccc = core.ComputeClient{}
	// //vcc = core.VirtualNetworkClient{}
	// lbc = loadbalancer.LoadBalancerClient{}
	// hcc = healthchecks.HealthChecksClient{}
	// dbc = database.DatabaseClient{}

	allRegionsResourceTags.Range(func(key, value interface{}) bool {
		backend.Logger.Info("client", "getResourceTags", "Resource tags got for region-"+key.(string))

		resourceTagsGot := value.(map[string][]string)

		// for first entry
		if len(allResourceTags) == 0 {
			allResourceTags = resourceTagsGot
			return true
		}

		// k will be tag key
		// values will be tag values
		for k, values := range resourceTagsGot {
			if _, ok := allResourceTags[k]; !ok {
				// when key not present
				allResourceTags[k] = values
				continue
			}

			// when key is already present
			for _, mn := range values {
				findIndex := sort.SearchStrings(allResourceTags[k], mn)
				if findIndex < len(allResourceTags[k]) && allResourceTags[k][findIndex] != mn {
					// not found, and insert in between
					allResourceTags[k] = append(allResourceTags[k][:findIndex+1], allResourceTags[k][findIndex:]...)
					allResourceTags[k][findIndex] = mn
				} else if findIndex == len(allResourceTags[k]) {
					// not found and insert at last
					allResourceTags[k] = append(allResourceTags[k], mn)
				}
			}
		}

		return true
	})

	for k, v := range allResourceTags {
		resourceTagsList = append(resourceTagsList, models.OCIResourceTags{
			Key:    k,
			Values: v,
		})
	}

	return resourceTagsList
}

// GetResourceGroups Returns all the resource groups associated with mentioned namespace under the compartment of mentioned tenancy
// API Operation: ListMetrics
// Permission Required: METRIC_INSPECT
// Links:
// https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/monitoringpolicyreference.htm
// https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/Metric/ListMetrics
// GetResourceGroups fetches the resource groups under a specified compartment and namespace.
// It first checks the cache for existing data and returns it if available. If not, it makes a request
// to the OCI Monitoring service to retrieve the resource groups and their associated metric names.
// The results are then cached for future use.
//
// Parameters:
//   - ctx: The context for the request.
//   - tenancyOCID: The OCID of the tenancy.
//   - compartmentOCID: The OCID of the compartment.
//   - region: The region to query. If set to constants.ALL_REGION, it queries all subscribed regions.
//   - namespace: The namespace to query.
//
// Returns:
//   - A slice of OCIMetricNamesWithResourceGroup, which contains the resource groups and their associated metric names.
func (o *OCIDatasource) GetResourceGroups(
	ctx context.Context,
	tenancyOCID string,
	compartmentOCID string,
	region string,
	namespace string) []models.OCIMetricNamesWithResourceGroup {
	backend.Logger.Error("client", "GetResourceGroups", "fetching the resource groups under compartment '"+compartmentOCID+"' for namespace '"+namespace+"'")

	// fetching from cache, if present
	cacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, namespace, "rgs"}, "-")

	if cachedResourceGroups, found := o.cache.Get(cacheKey); found {
		if rg, ok := cachedResourceGroups.([]models.OCIMetricNamesWithResourceGroup); ok {
			backend.Logger.Warn("client", "GetResourceGroups", "getting the data from cache")
			return rg
		}
	}

	var metricResourceGroups map[string][]string
	metricResourceGroupsList := []models.OCIMetricNamesWithResourceGroup{}
	takey := o.GetTenancyAccessKey(tenancyOCID)

	monitoringRequest := monitoring.ListMetricsRequest{
		CompartmentId:          common.String(compartmentOCID),
		CompartmentIdInSubtree: common.Bool(false),
		ListMetricsDetails: monitoring.ListMetricsDetails{
			GroupBy:   []string{"resourceGroup", "name"},
			Namespace: common.String(namespace),
		},
	}

	if len(compartmentOCID) == 0 {
		monitoringRequest.CompartmentId = common.String(tenancyOCID)
		monitoringRequest.CompartmentIdInSubtree = common.Bool(true)
	}

	if region == constants.ALL_REGION {
		metricResourceGroups = listMetricsMetadataFromAllRegion(
			ctx,
			o.cache,
			cacheKey,
			constants.FETCH_FOR_RESOURCE_GROUP,
			o.tenancyAccess[takey].monitoringClient,
			monitoringRequest,
			o.GetSubscribedRegions(ctx, tenancyOCID),
		)
	} else {
		metricResourceGroups = listMetricsMetadataPerRegion(
			ctx,
			o.cache,
			cacheKey,
			constants.FETCH_FOR_RESOURCE_GROUP,
			o.tenancyAccess[takey].monitoringClient,
			monitoringRequest,
		)
	}

	if len(metricResourceGroups) == 0 {
		backend.Logger.Error("client", "GetResourceGroups", "resource groups under compartment '"+compartmentOCID+"' for namespace '"+namespace+"' is empty")
		return nil
	} else {
		for k, v := range metricResourceGroups {
			metricResourceGroupsList = append(metricResourceGroupsList, models.OCIMetricNamesWithResourceGroup{
				ResourceGroup: k,
				MetricNames:   v,
			})
		}
	}

	// saving into the cache
	o.cache.SetWithTTL(cacheKey, metricResourceGroupsList, 1, 5*time.Minute)
	o.cache.Wait()

	return metricResourceGroupsList
}

// GetDimensions Returns all the dimensions associated with mentioned metric under the compartment of mentioned tenancy
// API Operation: ListMetrics
// Permission Required: METRIC_INSPECT
// Links:
// https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/monitoringpolicyreference.htm
// https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/Metric/ListMetrics
// GetDimensions retrieves the dimensions for a given metric in a specified compartment and namespace.
// It can be used to fetch dimensions for panels or labels based on the isLabel parameter.
//
// Parameters:
//   - ctx: The context for the request.
//   - tenancyOCID: The OCID of the tenancy.
//   - compartmentOCID: The OCID of the compartment.
//   - region: The region to query metrics from.
//   - namespace: The namespace of the metric.
//   - metricName: The name of the metric.
//   - isLabel: Optional boolean parameter to specify if dimensions are to be used as labels.
//
// Returns:
//   - A slice of OCIMetricDimensions containing the dimensions for the specified metric.
func (o *OCIDatasource) GetDimensions(
	ctx context.Context,
	tenancyOCID string,
	compartmentOCID string,
	region string,
	namespace string,
	metricName string,
	isLabel ...bool) []models.OCIMetricDimensions {

	// Check if we are using the function to retrieve Dimensions for panels or labels
	var DimensionUse string
	var cacheSubKey string
	if len(isLabel) > 0 {
		backend.Logger.Error("client", "GetDimensionsLabels", "fetching the dimension under compartment '"+compartmentOCID+"' for namespace '"+namespace+"' and metric '"+metricName+"' to be used as labels")
		DimensionUse = constants.FETCH_FOR_LABELDIMENSION
		cacheSubKey = "dslabel"
	} else {
		backend.Logger.Error("client", "GetDimensions", "fetching the dimension under compartment '"+compartmentOCID+"' for namespace '"+namespace+"' and metric '"+metricName+"'")
		DimensionUse = constants.FETCH_FOR_DIMENSION
		cacheSubKey = "ds"
	}

	// fetching from cache, if present
	cacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, namespace, metricName, cacheSubKey}, "-")
	if cachedDimensions, found := o.cache.Get(cacheKey); found {
		// This check avoids the type assertion and potential panic
		if _, ok := cachedDimensions.([]models.OCIMetricDimensions); ok {
			backend.Logger.Warn("client", "GetDimensions", "getting the data from cache")
			return cachedDimensions.([]models.OCIMetricDimensions)
		} else {
			backend.Logger.Warn("client.utils", "GetDimensions", "cannot use cached data -> "+cacheKey)
		}
	}

	var metricDimensions map[string][]string
	metricDimensionsList := []models.OCIMetricDimensions{}
	takey := o.GetTenancyAccessKey(tenancyOCID)

	if len(takey) == 0 {
		backend.Logger.Warn("client", "GetDimensions", "invalid takey")
		return nil
	}

	monitoringRequest := monitoring.ListMetricsRequest{
		CompartmentId:          common.String(compartmentOCID),
		CompartmentIdInSubtree: common.Bool(false),
		ListMetricsDetails: monitoring.ListMetricsDetails{
			Name:      common.String(metricName),
			Namespace: common.String(namespace),
		},
	}

	if len(compartmentOCID) == 0 {
		monitoringRequest.CompartmentId = common.String(tenancyOCID)
		monitoringRequest.CompartmentIdInSubtree = common.Bool(true)
	}

	if region == constants.ALL_REGION {
		metricDimensions = listMetricsMetadataFromAllRegion(
			ctx,
			o.cache,
			cacheKey,
			DimensionUse,
			o.tenancyAccess[takey].monitoringClient,
			monitoringRequest,
			o.GetSubscribedRegions(ctx, tenancyOCID),
		)
	} else {
		metricDimensions = listMetricsMetadataPerRegion(
			ctx,
			o.cache,
			cacheKey,
			DimensionUse,
			o.tenancyAccess[takey].monitoringClient,
			monitoringRequest,
		)
	}

	for k, v := range metricDimensions {
		metricDimensionsList = append(metricDimensionsList, models.OCIMetricDimensions{
			Key:    k,
			Values: v,
		})
	}

	// saving into the cache
	o.cache.SetWithTTL(cacheKey, metricDimensionsList, 1, 5*time.Minute)
	o.cache.Wait()

	return metricDimensionsList
}
