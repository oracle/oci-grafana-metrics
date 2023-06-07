package plugin

import (
	"context"
	"sort"
	"strings"
	"time"

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
	// var testResult bool
	// var errAllComp error

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
		o.tenancyAccess[key].metricsClient.SetRegion(string(reg))

		// Test Tenancy OCID first
		backend.Logger.Debug(key, "Testing Tenancy OCID", tenancyocid)
		listMetrics := monitoring.ListMetricsRequest{
			CompartmentId: &tenancyocid,
		}

		res, err := o.tenancyAccess[key].metricsClient.ListMetrics(ctx, listMetrics)
		if err != nil {
			backend.Logger.Debug(key, "SKIPPED", err)
		}
		status := res.RawResponse.StatusCode
		if status >= 200 && status < 300 {
			backend.Logger.Debug(key, "OK", status)
		} else {
			// backend.Logger.Debug(key, "SKIPPED", fmt.Sprintf("listMetrics on Tenancy %s did not work, testing compartments", tenancyocid))
			// comparts := o.GetCompartments(ctx, tenancyocid)

			// for _, v := range comparts {
			// 	backend.Logger.Debug(key, "Testing", v)
			// 	listMetrics := monitoring.ListMetricsRequest{
			// 		CompartmentId: common.String(v),
			// 	}

			// 	res, err := o.tenancyAccess[key].metricsClient.ListMetrics(ctx, listMetrics)
			// 	if err != nil {
			// 		backend.Logger.Debug(key, "FAILED", err)
			// 	}
			// 	status := res.RawResponse.StatusCode
			// 	if status >= 200 && status < 300 {
			// 		backend.Logger.Debug(key, "OK", status)
			// 		testResult = true
			// 		break
			// 	} else {
			// 		errAllComp = err
			// 		backend.Logger.Debug(key, "SKIPPED", status)
			// 	}
			// }
			// if testResult {
			// 	continue
			// } else {
			// 	backend.Logger.Debug(key, "FAILED", "listMetrics failed in each compartment")
			// 	return errors.Wrap(errAllComp, fmt.Sprintf("listMetrics failed in each Compartments in profile %s", key))
			// }
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

	tenancyocid, tenancyErr := o.tenancyAccess[takey].config.TenancyOCID()
	if tenancyErr != nil {
		return nil
	}
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

	tenancymode := o.settings.TenancyMode

	region, regErr := o.tenancyAccess[takey].config.Region()
	if regErr != nil {
		return nil
	}
	reg := common.StringToRegion(region)
	o.tenancyAccess[takey].metricsClient.SetRegion(string(reg))

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
		return nil
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

	return compartmentList
}

// // GetNamespaceWithMetricNames Returns all the namespaces with associated metrics under the compartment of mentioned tenancy
// // API Operation: ListMetrics
// // Permission Required: METRIC_INSPECT
// // Links:
// // https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/monitoringpolicyreference.htm
// // https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/Metric/ListMetrics
// func (oc *OCIClients) GetNamespaceWithMetricNames(
// 	ctx context.Context,
// 	tenancyOCID string,
// 	compartmentOCID string,
// 	region string) []models.OCIMetricNamesWithNamespace {
// 	backend.Logger.Debug("client", "GetNamespaceWithMetricNames", "fetching the metric names along with namespaces under compartment: "+compartmentOCID)

// 	// fetching from cache, if present
// 	cacheKey := strings.Join([]string{tenancyOCID, compartmentOCID, region, "nss"}, "-")
// 	if cachedMetricNamesWithNamespaces, found := oc.cache.Get(cacheKey); found {
// 		backend.Logger.Warn("client", "GetNamespaceWithMetricNames", "getting the data from cache")
// 		return cachedMetricNamesWithNamespaces.([]models.OCIMetricNamesWithNamespace)
// 	}

// 	// calling the api if not present in cache
// 	var namespaceWithMetricNames map[string][]string
// 	namespaceWithMetricNamesList := []models.OCIMetricNamesWithNamespace{}

// 	client := oc.GetOciClient(tenancyOCID)
// 	if client == nil {
// 		return namespaceWithMetricNamesList
// 	}

// 	monitoringRequest := monitoring.ListMetricsRequest{
// 		CompartmentId:          common.String(compartmentOCID),
// 		CompartmentIdInSubtree: common.Bool(false),
// 		ListMetricsDetails: monitoring.ListMetricsDetails{
// 			GroupBy:   []string{"namespace", "name"},
// 			SortBy:    monitoring.ListMetricsDetailsSortByNamespace,
// 			SortOrder: monitoring.ListMetricsDetailsSortOrderAsc,
// 		},
// 	}

// 	// when search is wide along the tenancy
// 	if len(compartmentOCID) == 0 {
// 		monitoringRequest.CompartmentId = common.String(tenancyOCID)
// 		monitoringRequest.CompartmentIdInSubtree = common.Bool(true)
// 	}

// 	// when user wants to fetch everything for all subscribed regions
// 	if region == constants.ALL_REGION {
// 		namespaceWithMetricNames = listMetricsMetadataFromAllRegion(
// 			ctx,
// 			oc.cache,
// 			cacheKey,
// 			constants.FETCH_FOR_NAMESPACE,
// 			client.monitoringClient,
// 			monitoringRequest,
// 			oc.GetSubscribedRegions(ctx, tenancyOCID),
// 		)
// 	} else {
// 		if region != "" {
// 			client.monitoringClient.SetRegion(region)
// 		}
// 		namespaceWithMetricNames = listMetricsMetadataPerRegion(
// 			ctx,
// 			oc.cache,
// 			cacheKey,
// 			constants.FETCH_FOR_NAMESPACE,
// 			client.monitoringClient,
// 			monitoringRequest,
// 		)
// 	}

// 	// preparing for frontend
// 	for k, v := range namespaceWithMetricNames {
// 		namespaceWithMetricNamesList = append(namespaceWithMetricNamesList, models.OCIMetricNamesWithNamespace{
// 			Namespace:   k,
// 			MetricNames: v,
// 		})
// 	}

// 	// saving into the cache
// 	oc.cache.SetWithTTL(cacheKey, namespaceWithMetricNamesList, 1, 5*time.Minute)
// 	oc.cache.Wait()

// 	return namespaceWithMetricNamesList
// }
