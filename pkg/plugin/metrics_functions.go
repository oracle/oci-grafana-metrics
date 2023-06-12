package plugin

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
	"github.com/pkg/errors"
)

// TestConnectivity Check the OCI data source test request in Grafana's Datasource configuration UI.
func (o *OCIDatasource) TestConnectivity(ctx context.Context) error {
	backend.Logger.Debug("client", "TestConnectivity", "testing the OCI connectivity")

	var reg common.Region
	var testResult bool
	var errAllComp error

	// tenv := o.settings.Environment
	// tmode := o.settings.TenancyMode

	for key, _ := range o.tenancyAccess {
		// testResult = false

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

		res, err := o.tenancyAccess[key].monitoringClient.ListMetrics(ctx, listMetrics)
		if err != nil {
			backend.Logger.Debug(key, "SKIPPED", err)
		}
		status := res.RawResponse.StatusCode
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

	backend.Logger.Debug("client", "GetSubscribedRegionstakey", "fetching the subscribed region for tenancy takey: "+takey)

	tenancyocid, tenancyErr := o.tenancyAccess[takey].config.TenancyOCID()
	if tenancyErr != nil {
		return nil
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

	// saving into the cache
	o.cache.SetWithTTL(cacheKey, namespaceWithMetricNamesList, 1, 5*time.Minute)
	o.cache.Wait()

	return namespaceWithMetricNamesList
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
