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

// TestConnectivity Check the OCI data source test request in Grafana's Datasource configuration UI.
func (o *OCIDatasource) TestConnectivity(ctx context.Context) error {
	backend.Logger.Debug("client", "TestConnectivity", "testing the OCI connectivity")

	var reg common.Region
	var testResult bool
	var errAllComp error

	// tenv := o.settings.Environment
	// tmode := o.settings.TenancyMode

	for key, _ := range o.tenancyAccess {
		testResult = false

		// if tmode == "multitenancy" && tenv == "oci-instance" {
		// 	return errors.New("Multitenancy mode using instance principals is not implemented yet.")
		// }
		tenancyocid, tenancyErr := o.tenancyAccess[key].config.TenancyOCID()
		if tenancyErr != nil {
			return errors.Wrap(tenancyErr, "error fetching TenancyOCID")
		}

		regio, regErr := o.tenancyAccess[key].config.Region()
		if regErr != nil {
			return errors.Wrap(regErr, "error fetching Region")
		}
		reg = common.StringToRegion(regio)
		o.tenancyAccess[key].monitoringClient.SetRegion(string(reg))

		// Test Tenancy OCID first
		backend.Logger.Debug(key, "Testing Tenancy OCID", tenancyocid)
		listMetrics := monitoring.ListMetricsRequest{
			CompartmentId: &tenancyocid,
		}

		var status int
		res, err := o.tenancyAccess[key].monitoringClient.ListMetrics(ctx, listMetrics)
		if err != nil {
			backend.Logger.Debug(key, "SKIPPED", err)
		} else {
			status = res.RawResponse.StatusCode
		}
		if status >= 200 && status < 300 {
			backend.Logger.Debug(key, "OK", status)
		} else {
			backend.Logger.Debug(key, "SKIPPED", fmt.Sprintf("listMetrics on Tenancy %s did not work, testing compartments", tenancyocid))
			comparts := o.GetCompartments(ctx, tenancyocid)

			for _, v := range comparts {
				tocid := v.OCID
				backend.Logger.Debug(key, "Testing", tocid)
				listMetrics := monitoring.ListMetricsRequest{
					CompartmentId: common.String(tocid),
				}

				res, err := o.tenancyAccess[key].monitoringClient.ListMetrics(ctx, listMetrics)
				if err != nil {
					backend.Logger.Debug(key, "FAILED", err)
				}
				status := res.RawResponse.StatusCode
				if status >= 200 && status < 300 {
					backend.Logger.Debug(key, "OK", status)
					testResult = true
					break
				} else {
					errAllComp = err
					backend.Logger.Debug(key, "SKIPPED", status)
				}
			}
			if testResult {
				continue
			} else {
				backend.Logger.Debug(key, "FAILED", "listMetrics failed in each compartment")
				return errors.Wrap(errAllComp, fmt.Sprintf("listMetrics failed in each Compartments in profile %s", key))
			}
		}

	}
	return nil

}

/*
Function generates an array  containing OCI tenancy list in the following format:
<Label/TenancyOCID>
*/
func (o *OCIDatasource) GetTenancies(ctx context.Context) []models.OCIResource {
	backend.Logger.Debug("client", "GetTenancies", "fetching the tenancies")

	tenancyList := []models.OCIResource{}
	for key, _ := range o.tenancyAccess {
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
func (o *OCIDatasource) GetSubscribedRegions(ctx context.Context, tenancyOCID string) []string {
	backend.Logger.Debug("client", "GetSubscribedRegions", "fetching the subscribed region for tenancy: "+tenancyOCID)

	var subscribedRegions []string
	takey := o.GetTenancyAccessKey(tenancyOCID)
	tenancymode := o.settings.TenancyMode
	var tenancyocid string
	var tenancyErr error

	backend.Logger.Debug("client", "GetSubscribedRegionstakey", "fetching the subscribed region for tenancy takey: "+takey)

	if tenancymode == "multitenancy" {
		if len(takey) <= 0 || takey == NoTenancy {
			o.logger.Error("Unable to get Multi-tenancy OCID")
			return nil
		}
		res := strings.Split(takey, "/")
		tenancyocid = res[1]
	} else {
		tenancyocid, tenancyErr = o.tenancyAccess[takey].config.TenancyOCID()
		if tenancyErr != nil {
			return nil
		}
	}
	backend.Logger.Debug("client", "GetSubscribedRegionstakey", "fetching the subscribed region for tenancy OCID: "+*common.String(tenancyocid))

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
			backend.Logger.Debug("client", "GetSubscribedRegionstakey", "fetching the subscribed region for regioname: "+*item.RegionName)
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
func (o *OCIDatasource) GetCompartments(ctx context.Context, tenancyOCID string) []models.OCIResource {
	backend.Logger.Debug("client", "GetCompartments", "fetching the sub-compartments for tenancy: "+tenancyOCID)

	takey := o.GetTenancyAccessKey(tenancyOCID)
	var tenancyocid string
	var tenancyErr error
	backend.Logger.Debug("client", "GetCompartmentstakey", "fetching the subscribed region for tenancy takey: "+takey)
	backend.Logger.Debug("client", "GetCompartmentstakey", "fetching the subscribed region for tenancy tenancyOCID: "+tenancyOCID)
	tenancymode := o.settings.TenancyMode

	region, regErr := o.tenancyAccess[takey].config.Region()
	if regErr != nil {
		backend.Logger.Debug("client", "GetCompartments", "error retrieving default region")
		return nil
	}
	reg := common.StringToRegion(region)
	o.tenancyAccess[takey].monitoringClient.SetRegion(string(reg))

	if tenancymode == "multitenancy" {
		if len(takey) <= 0 || takey == NoTenancy {
			o.logger.Error("Unable to get Multi-tenancy OCID")
			return nil
		}
		res := strings.Split(takey, "/")
		tenancyocid = res[1]
	} else {
		tenancyocid, tenancyErr = o.tenancyAccess[takey].config.TenancyOCID()
		if tenancyErr != nil {
			return nil
		}
	}

	backend.Logger.Debug("client", "GetCompartmentstakey2", "fetching the subscribed region for tenancy tenancyOCID: "+tenancyOCID)

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
		backend.Logger.Debug("client", "GetCompartments", "error in GetTenancy")
		return nil
	}
	backend.Logger.Debug("client", "GetCompartmentstakey3", "fetching the subscribed region for tenancy tenancyOCID: "+tenancyOCID)

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
				AccessLevel:            identity.ListCompartmentsAccessLevelAny,
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

	backend.Logger.Debug("client", "GetCompartmentstakey4", "fetching the subscribed region for tenancy tenancyOCID: "+*resp.Name)

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

	if len(compartmentList) > 1 {
		compartmentList = append(compartmentList, models.OCIResource{
			Name: constants.ALL_COMPARTMENT,
			OCID: "",
		})
	}

	// sorting based on compartment name
	sort.SliceStable(compartmentList, func(i, j int) bool {
		return compartmentList[i].Name < compartmentList[j].Name
	})

	// saving in the cache
	o.cache.SetWithTTL(cacheKey, compartmentList, 1, 15*time.Minute)
	o.cache.Wait()

	backend.Logger.Debug("client", "GetCompartmentstakey5", "fetching the subscribed region for tenancy tenancyOCID: "+*resp.Name)

	return compartmentList
}

// GetNamespaceWithMetricNames Returns all the namespaces with associated metrics under the compartment of mentioned tenancy
// API Operation: ListMetrics
// Permission Required: METRIC_INSPECT
// Links:
// https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/monitoringpolicyreference.htm
// https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/Metric/ListMetrics
func (o *OCIDatasource) GetNamespaceWithMetricNames(
	ctx context.Context,
	tenancyOCID string,
	compartmentOCID string,
	region string) []models.OCIMetricNamesWithNamespace {
	backend.Logger.Debug("client", "GetNamespaceWithMetricNames", "fetching the metric names along with namespaces under compartment: "+compartmentOCID)

	takey := o.GetTenancyAccessKey(tenancyOCID)
	// fetching from cache, if present
	cacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, "nss"}, "-")
	if cachedMetricNamesWithNamespaces, found := o.cache.Get(cacheKey); found {
		backend.Logger.Warn("client", "GetNamespaceWithMetricNames", "getting the data from cache")
		return cachedMetricNamesWithNamespaces.([]models.OCIMetricNamesWithNamespace)
	}

	// calling the api if not present in cache
	var namespaceWithMetricNames map[string][]string
	namespaceWithMetricNamesList := []models.OCIMetricNamesWithNamespace{}

	// client := oc.GetOciClient(tenancyOCID)
	// if client == nil {
	// 	return namespaceWithMetricNamesList
	// }

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
		if region != "" {
			o.tenancyAccess[takey].monitoringClient.SetRegion(region)
		}
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

// GetMetricDataPoints Returns metric datapoints
// API Operation: SummarizeMetricsData
// Permission Required: METRIC_INSPECT and METRIC_READ
// Links:
// https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/monitoringpolicyreference.htm
// https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/MetricData/SummarizeMetricsData
func (o *OCIDatasource) GetMetricDataPoints(ctx context.Context, requestParams models.MetricsDataRequest, tenancyOCID string) ([]time.Time, []models.OCIMetricDataPoints) {
	backend.Logger.Debug("client", "GetMetricDataPoints", "fetching the metrics datapoints under compartment '"+requestParams.CompartmentOCID+"' for query '"+requestParams.QueryText+"'")

	times := []time.Time{}
	dataValuesWithTime := map[common.SDKTime][]float64{}
	dataPointsWithResourceSerialNo := map[int]models.OCIMetricDataPoints{}
	dataPoints := []models.OCIMetricDataPoints{}
	resourceIDsPerTag := map[string]map[string]struct{}{}

	selectedTags := requestParams.TagsValues
	selectedDimensions := requestParams.DimensionValues
	selectedLegendFormat := requestParams.LegendFormat
	o.logger.Debug("selectedLegendFormat", "selectedLegendFormat", selectedLegendFormat)

	takey := o.GetTenancyAccessKey(tenancyOCID)

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

	// to search for all copartments
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
	for _, subscribedRegion := range subscribedRegions {
		if subscribedRegion != constants.ALL_REGION {
			wg.Add(1)
			go func(mc monitoring.MonitoringClient, sRegion string) {
				defer wg.Done()

				mc.SetRegion(sRegion)
				resp, err := mc.SummarizeMetricsData(ctx, metricsDataRequest)
				if err != nil {
					backend.Logger.Error("client", "GetMetricDataPoints", err)
					return
				}

				if len(resp.Items) > 0 {
					// fetching the resource labels
					var rl map[string]map[string]string

					cachedResourceLabels := o.fetchFromCache(
						ctx,
						requestParams.TenancyOCID,
						requestParams.CompartmentOCID,
						requestParams.CompartmentName,
						sRegion,
						requestParams.Namespace,
						"resource_labels",
					)

					rl = cachedResourceLabels.(map[string]map[string]string)

					// storing the data to calculate later
					allRegionsMetricsDataPoint.Store(sRegion, metricDataBank{
						dataPoints:     resp.Items,
						resourceLabels: rl,
					})
				}
			}(o.tenancyAccess[takey].monitoringClient, subscribedRegion)
		}
	}
	wg.Wait()

	resourcesFetched := 0

	allRegionsMetricsDataPoint.Range(func(key, value interface{}) bool {
		regionInUse := key.(string)

		backend.Logger.Info("client", "GetMetricDataPoints", "Metric datapoints got for region-"+regionInUse)

		// get the selected tags
		if len(selectedTags) != 0 {
			cachedResourceNamesPerTag := o.fetchFromCache(
				ctx,
				requestParams.TenancyOCID,
				requestParams.CompartmentOCID,
				requestParams.CompartmentName,
				regionInUse,
				requestParams.Namespace,
				constants.CACHE_KEY_RESOURCE_IDS_PER_TAG,
			)

			resourceIDsPerTag = cachedResourceNamesPerTag.(map[string]map[string]struct{})
		}

		metricData := value.(metricDataBank)

		for _, metricDataItem := range metricData.dataPoints {
			found := false

			uniqueDataID, resourceDisplayName, extraUniqueID, rIDPresent := getUniqueIdsForLabels(requestParams.Namespace, metricDataItem.Dimensions)

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

					// adjustment for previous non-existance values
					// when the time comes in later data points
					// dataValuesWithTime[t] = []float64{0.0}
					// for i := 2; i < resourcesFetched; i++ {
					// 	dataValuesWithTime[t] = append(dataValuesWithTime[t], 0.0)
					// }
					// dataValuesWithTime[t] = append(dataValuesWithTime[t], v)

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
			// tenancyName := oc.tenanciesMap[requestParams.TenancyOCID]
			// if tenancyName == constants.DEFAULT_PROFILE {
			// 	tenancyName = oc.baseTenancyName
			// }

			// to get the resource labels
			labelKey := uniqueDataID + extraUniqueID
			if strings.Contains(resourceDisplayName, "ocid") {
				resourceDisplayName = metricData.resourceLabels[labelKey]["resource_name"]
			}

			// adding the selected dimensions as labels
			labelsToAdd := addSelectedValuesLabels(metricData.resourceLabels[labelKey], selectedDimensions)
			// adding the selected tags as labels
			labelsToAdd = addSelectedValuesLabels(labelsToAdd, selectedTags)

			// preparing the metric data to display
			dataPointsWithResourceSerialNo[resourcesFetched-1] = models.OCIMetricDataPoints{
				TenancyName:  tenancyName,
				Region:       regionInUse,
				MetricName:   *metricDataItem.Name,
				ResourceName: resourceDisplayName,
				UniqueDataID: uniqueDataID,
				Labels:       labelsToAdd,
			}

			// // adding cmdb data as labels
			// for ocid, cmdbData := range oc.cmdbData[tenancyName] {
			// 	if ocid != uniqueDataID {
			// 		// when there is no data for the resource ocid
			// 		continue
			// 	}

			// 	dp := dataPointsWithResourceSerialNo[resourcesFetched-1]

			// 	// adding to the existing labels
			// 	for k, v := range cmdbData {
			// 		dp.Labels[k] = v
			// 	}

			// 	dataPointsWithResourceSerialNo[resourcesFetched-1] = dp
			// }
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

		// for i := 0; i < resourcesFetched-len(dvs); i++ {
		// 	dataValuesWithTime[t] = append(dataValuesWithTime[t], 0.0)
		// }

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

	return times, dataPoints
}

// fetchFromCache will fetch value from cache and if it not present it will fetch via api and store to cache and return
func (o *OCIDatasource) fetchFromCache(ctx context.Context, tenancyOCID string, compartmentOCID string, compartmentName string, region string, namespace string, suffix string) interface{} {
	backend.Logger.Debug("client", "fetchFromCache", "fetching from cache")

	labelCacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, namespace, suffix}, "-")

	if _, found := o.cache.Get(labelCacheKey); !found {
		o.GetTags(ctx, tenancyOCID, compartmentOCID, compartmentName, region, namespace)
	}

	cachedResource, _ := o.cache.Get(labelCacheKey)
	return cachedResource
}

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
	backend.Logger.Debug("client", "GetTags", "fetching the tags for namespace '"+namespace+"'")

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

	compartments := []models.OCIResource{}
	if len(compartmentOCID) == 0 {
		compartments = append(compartments, o.GetCompartments(ctx, tenancyOCID)...)
	} else {
		compartments = append(compartments, models.OCIResource{
			Name: compartmentName,
			OCID: compartmentOCID,
		})
	}

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
func (o *OCIDatasource) GetResourceGroups(
	ctx context.Context,
	tenancyOCID string,
	compartmentOCID string,
	region string,
	namespace string) []models.OCIMetricNamesWithResourceGroup {
	backend.Logger.Debug("client", "GetResourceGroups", "fetching the resource groups under compartment '"+compartmentOCID+"' for namespace '"+namespace+"'")

	// fetching from cache, if present
	cacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, namespace, "rgs"}, "-")
	if cachedResourceGroups, found := o.cache.Get(cacheKey); found {
		backend.Logger.Warn("client", "GetResourceGroups", "getting the data from cache")
		return cachedResourceGroups.([]models.OCIMetricNamesWithResourceGroup)
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
		if region != "" {
			o.tenancyAccess[takey].monitoringClient.SetRegion(region)
		}

		metricResourceGroups = listMetricsMetadataPerRegion(
			ctx,
			o.cache,
			cacheKey,
			constants.FETCH_FOR_RESOURCE_GROUP,
			o.tenancyAccess[takey].monitoringClient,
			monitoringRequest,
		)
	}

	reqDetails := monitoring.ListMetricsDetails{}
	reqDetails.Namespace = common.String(namespace)
	reqDetails.GroupBy = []string{"resourceGroup", "name"}
	items, err := o.searchHelper(ctx, region, compartmentOCID, reqDetails, takey)
	if err != nil {
		return nil
	}

	if len(metricResourceGroups) == 0 {
		var arca []string
		for _, item := range items {
			alfa := *(item.Name)
			arca = append(arca, alfa)
			backend.Logger.Debug("client", "GetResourceGroups k", "alfa the resource groups under compartment '"+compartmentOCID+"' for namespace '"+namespace+"' "+alfa)
		}
		metricResourceGroupsList = append(metricResourceGroupsList, models.OCIMetricNamesWithResourceGroup{
			ResourceGroup: constants.DEFAULT_RESOURCE_GROUP,
			MetricNames:   arca,
		})
	} else {
		for k, v := range metricResourceGroups {
			metricResourceGroupsList = append(metricResourceGroupsList, models.OCIMetricNamesWithResourceGroup{
				ResourceGroup: k,
				MetricNames:   v,
			})
		}
		if len(metricResourceGroupsList) > 0 {
			metricResourceGroupsList = append(metricResourceGroupsList, models.OCIMetricNamesWithResourceGroup{
				ResourceGroup: constants.DEFAULT_RESOURCE_GROUP,
				MetricNames:   []string{},
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
		backend.Logger.Debug("client", "GetDimensionsLabels", "fetching the dimension under compartment '"+compartmentOCID+"' for namespace '"+namespace+"' and metric '"+metricName+"' to be used as labels")
		DimensionUse = constants.FETCH_FOR_LABELDIMENSION
		cacheSubKey = "dslabel"
	} else {
		backend.Logger.Debug("client", "GetDimensions", "fetching the dimension under compartment '"+compartmentOCID+"' for namespace '"+namespace+"' and metric '"+metricName+"'")
		DimensionUse = constants.FETCH_FOR_DIMENSION
		cacheSubKey = "ds"
	}

	// fetching from cache, if present
	cacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, namespace, metricName, cacheSubKey}, "-")
	if cachedDimensions, found := o.cache.Get(cacheKey); found {
		backend.Logger.Warn("client", "GetDimensions", "getting the data from cache")
		return cachedDimensions.([]models.OCIMetricDimensions)
	}

	var metricDimensions map[string][]string
	metricDimensionsList := []models.OCIMetricDimensions{}
	takey := o.GetTenancyAccessKey(tenancyOCID)

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
		if region != "" {
			o.tenancyAccess[takey].monitoringClient.SetRegion(region)
		}

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

func (o *OCIDatasource) searchHelper(ctx context.Context, region, compartment string, metricDetails monitoring.ListMetricsDetails, takey string) ([]monitoring.Metric, error) {
	var items []monitoring.Metric
	var page *string

	pageNumber := 0
	for {
		reg := common.StringToRegion(region)
		o.tenancyAccess[takey].monitoringClient.SetRegion(string(reg))
		res, err := o.tenancyAccess[takey].monitoringClient.ListMetrics(ctx, monitoring.ListMetricsRequest{
			CompartmentId:      common.String(compartment),
			ListMetricsDetails: metricDetails,
			Page:               page,
		})

		if err != nil {
			return nil, errors.Wrap(err, "list metrics failed")
		}
		items = append(items, res.Items...)
		// Only 0 - n-1  pages are to be fetched, as indexing starts from 0 (for page number
		if res.OpcNextPage == nil || pageNumber >= MaxPagesToFetch {
			break
		}

		page = res.OpcNextPage
		pageNumber++
	}
	return items, nil
}
