// Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/davecgh/go-spew/spew"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/pkg/errors"
)

const MaxPagesToFetch = 20
const SingleTenancyKey = "DEFAULT/"

var (
	cacheRefreshTime = time.Minute // how often to refresh our compartmentID cache
	re               = regexp.MustCompile(`(?m)\w+Name`)
)

// OCIDatasource - pulls in data from telemtry/various oci apis
// type OCIDatasource struct {
// 	metricsClient    monitoring.MonitoringClient
// 	identityClient   identity.IdentityClient
// 	config           common.ConfigurationProvider
// 	logger           log.Logger
// 	nameToOCID       map[string]string
// 	timeCacheUpdated time.Time
// }

type OCIDatasource struct {
	tenancyAccess    map[string]*TenancyAccess
	logger           log.Logger
	nameToOCID       map[string]string
	timeCacheUpdated time.Time
}

// NewOCIDatasource - constructor
func NewOCIDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &OCIDatasource{
		tenancyAccess: make(map[string]*TenancyAccess),
		logger:        log.DefaultLogger,
		nameToOCID:    make(map[string]string),
	}, nil
}

type TenancyAccess struct {
	metricsClient  monitoring.MonitoringClient
	identityClient identity.IdentityClient
	config         common.ConfigurationProvider
}

// GrafanaOCIRequest - Query Request comning in from the front end
type GrafanaOCIRequest struct {
	GrafanaCommonRequest
	Query         string
	Resolution    string
	Namespace     string
	TenancyConfig string // the actual tenancy with the format <configfile entry name/tenancyOCID>
	ResourceGroup string
	LegendFormat  string
}

// GrafanaSearchRequest incoming request body for search requests
type GrafanaSearchRequest struct {
	GrafanaCommonRequest
	Metric        string `json:"metric,omitempty"`
	Namespace     string
	TenancyConfig string // the actual tenancy with the format <configfile entry name/tenancyOCID>
	ResourceGroup string
}

// GrafanaCommonRequest - captures the common parts of the search and metricsRequests
type GrafanaCommonRequest struct {
	Compartment   string
	Environment   string
	TenancyMode   string
	QueryType     string
	Region        string
	TenancyConfig string // the actual tenancy with the format <configfile entry name/tenancyOCID>
	TenancyOCID   string `json:"tenancyOCID"`
}

// Query - Determine what kind of query we're making
func (o *OCIDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	var ts GrafanaCommonRequest
	var takey string

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	queryType := ts.QueryType

	o.logger.Debug("QueryData")
	o.logger.Debug(ts.Environment)
	o.logger.Debug(ts.TenancyMode)
	o.logger.Debug(ts.Region)
	o.logger.Debug(ts.TenancyConfig)

	// uncomment to use the single OCI login method
	// if len(o.tenancyAccess) == 0 {
	// uncomment to force OCI login at every query
	if true {

		err := o.getConfigProvider(ts.Environment, ts.TenancyMode)
		if err != nil {
			return nil, errors.Wrap(err, "broken environment")
		}
	}

	if ts.TenancyConfig != "NoTenancyConfig" && ts.TenancyConfig != "" {
		takey = ts.TenancyConfig
	} else {
		takey = SingleTenancyKey
	}

	o.logger.Debug(takey)
	o.logger.Debug(queryType)
	o.logger.Debug("/QueryData")

	switch queryType {
	case "compartments":
		return o.compartmentsResponse(ctx, req, takey)
	case "dimensions":
		return o.dimensionResponse(ctx, req, takey)
	case "namespaces":
		return o.namespaceResponse(ctx, req, takey)
	case "resourcegroups":
		return o.resourcegroupsResponse(ctx, req, takey)
	case "regions":
		return o.regionsResponse(ctx, req, takey)
	case "tenancyconfig":
		return o.tenancyConfigResponse(ctx, req)
	case "search":
		return o.searchResponse(ctx, req, takey)
	case "test":
		return o.testResponse(ctx, req)
	default:
		return o.queryResponse(ctx, req)
	}
}

func (o *OCIDatasource) testResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	var ts GrafanaCommonRequest
	var tenancyocid string
	var tenancyErr error

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	regions, _ := OCIConfigParser("regions")
	rr := 0
	reg := common.StringToRegion(ts.Region)

	for key, _ := range o.tenancyAccess { // Order not specified
		if ts.TenancyMode == "multitenancy" {
			tenancyocid, tenancyErr = o.tenancyAccess[key].config.TenancyOCID()
			if tenancyErr != nil {
				return nil, errors.Wrap(tenancyErr, "error fetching TenancyOCID")
			}
			reg = common.StringToRegion(regions[rr])
			rr++
		} else {
			tenancyocid = ts.TenancyOCID
		}
		listMetrics := monitoring.ListMetricsRequest{
			CompartmentId: common.String(tenancyocid),
		}
		o.tenancyAccess[key].metricsClient.SetRegion(string(reg))
		res, err := o.tenancyAccess[key].metricsClient.ListMetrics(ctx, listMetrics)
		if err != nil {
			return &backend.QueryDataResponse{}, err
		}
		status := res.RawResponse.StatusCode
		if status >= 200 && status < 300 {
			// return &backend.QueryDataResponse{}, nil
			o.logger.Error(key, "OK", status)
		} else {
			return nil, errors.Wrap(err, fmt.Sprintf("list metrircs failed %s %d", spew.Sdump(res), status))
		}
	}
	return &backend.QueryDataResponse{}, nil
}

func (o *OCIDatasource) dimensionResponse(ctx context.Context, req *backend.QueryDataRequest, takey string) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaSearchRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		reqDetails := monitoring.ListMetricsDetails{}
		reqDetails.Namespace = common.String(ts.Namespace)
		if ts.ResourceGroup != "NoResourceGroup" {
			reqDetails.ResourceGroup = common.String(ts.ResourceGroup)
		}
		reqDetails.Name = common.String(ts.Metric)
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails, takey)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint("list metrircs failed", spew.Sdump(reqDetails)))
		}

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))

		for _, item := range items {
			for dimension, value := range item.Dimensions {
				frame.AppendRow(fmt.Sprintf("%s=%s", dimension, value))
			}
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
}

func (o *OCIDatasource) namespaceResponse(ctx context.Context, req *backend.QueryDataRequest, takey string) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaSearchRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		reqDetails := monitoring.ListMetricsDetails{}
		reqDetails.GroupBy = []string{"namespace"}
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails, takey)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint("list metrircs failed", spew.Sdump(reqDetails)))
		}

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))
		for _, item := range items {
			frame.AppendRow(*(item.Namespace))
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
}

func (o *OCIDatasource) resourcegroupsResponse(ctx context.Context, req *backend.QueryDataRequest, takey string) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaSearchRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		reqDetails := monitoring.ListMetricsDetails{}
		reqDetails.Namespace = common.String(ts.Namespace)
		reqDetails.GroupBy = []string{"resourceGroup"}
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails, takey)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint("list metrircs failed", spew.Sdump(reqDetails)))
		}

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))

		frame.AppendRow(*(common.String("NoResourceGroup")))
		for _, item := range items {
			frame.AppendRow(*(item.ResourceGroup))
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
}

func (o *OCIDatasource) getConfigProvider(environment string, tenancymode string) error {

	o.logger.Debug("getConfigProvider")
	o.logger.Debug(environment)
	o.logger.Debug(tenancymode)
	switch environment {
	case "local":
		if tenancymode == "multitenancy" {
			ociconfigs, _ := OCIConfigParser("ociconfigs")
			for _, ociconfig := range ociconfigs {
				var configProvider common.ConfigurationProvider
				configProvider = common.CustomProfileConfigProvider("", ociconfig)
				metricsClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
				if err != nil {
					o.logger.Error("Error with config:" + ociconfig)
					return errors.New(fmt.Sprint("error with client", spew.Sdump(configProvider), err.Error()))
				}
				identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
				if err != nil {
					o.logger.Error("Error creating identity client", "error", err)
					return errors.Wrap(err, "Error creating identity client")
				}
				tenancyocid, err := configProvider.TenancyOCID()
				if err != nil {
					return errors.New(fmt.Sprint("error with TenancyOCID", spew.Sdump(configProvider), err.Error()))
				}
				o.tenancyAccess[ociconfig+"/"+tenancyocid] = &TenancyAccess{metricsClient, identityClient, configProvider}

				// o.tenancyAccess[ociconfig].identityClient = identityClient
				// o.tenancyAccess[ociconfig].metricsClient = metricsClient
				// o.tenancyAccess[ociconfig].config = configProvider

			}
			for key, _ := range o.tenancyAccess {
				o.logger.Debug(string(key))
			}
			return nil
		} else {
			var configProvider common.ConfigurationProvider
			configProvider = common.DefaultConfigProvider()
			metricsClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
			if err != nil {
				o.logger.Error("Error with config:" + SingleTenancyKey)
				return errors.New(fmt.Sprint("error with client", spew.Sdump(configProvider), err.Error()))
			}
			identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
			if err != nil {
				o.logger.Error("Error creating identity client", "error", err)
				return errors.Wrap(err, "Error creating identity client")
			}
			o.tenancyAccess[SingleTenancyKey] = &TenancyAccess{metricsClient, identityClient, configProvider}
			return nil
		}
	case "OCI Instance":
		var configProvider common.ConfigurationProvider
		configProvider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return errors.New(fmt.Sprint("error with instance principals", spew.Sdump(configProvider), err.Error()))
		}
		metricsClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
		if err != nil {
			o.logger.Error("Error with config:" + SingleTenancyKey)
			return errors.New(fmt.Sprint("error with client", spew.Sdump(configProvider), err.Error()))
		}
		identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
		if err != nil {
			o.logger.Error("Error creating identity client", "error", err)
			return errors.Wrap(err, "Error creating identity client")
		}
		o.tenancyAccess[SingleTenancyKey] = &TenancyAccess{metricsClient, identityClient, configProvider}
		return nil

	default:
		return errors.New("unknown environment type")
	}
}

func (o *OCIDatasource) searchResponse(ctx context.Context, req *backend.QueryDataRequest, takey string) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaSearchRequest

		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		reqDetails := monitoring.ListMetricsDetails{}
		// Group by is needed to get all  metrics without missing any as it is limited by the max pages
		reqDetails.GroupBy = []string{"name"}
		reqDetails.Namespace = common.String(ts.Namespace)
		if ts.ResourceGroup != "NoResourceGroup" {
			reqDetails.ResourceGroup = common.String(ts.ResourceGroup)
		}

		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails, takey)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint("list metrircs failed", spew.Sdump(reqDetails)))
		}

		metricCache := make(map[string]bool)

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))
		for _, item := range items {
			if _, ok := metricCache[*(item.Name)]; !ok {
				frame.AppendRow(*(item.Name))
				metricCache[*(item.Name)] = true
			}
		}
		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}

	return resp, nil
}

func (o *OCIDatasource) searchHelper(ctx context.Context, region, compartment string, metricDetails monitoring.ListMetricsDetails, takey string) ([]monitoring.Metric, error) {
	var items []monitoring.Metric
	var page *string

	pageNumber := 0
	for {
		reg := common.StringToRegion(region)
		o.tenancyAccess[takey].metricsClient.SetRegion(string(reg))
		res, err := o.tenancyAccess[takey].metricsClient.ListMetrics(ctx, monitoring.ListMetricsRequest{
			CompartmentId:      common.String(compartment),
			ListMetricsDetails: metricDetails,
			Page:               page,
		})

		if err != nil {
			return nil, errors.Wrap(err, "list metrircs failed")
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

func (o *OCIDatasource) compartmentsResponse(ctx context.Context, req *backend.QueryDataRequest, takey string) (*backend.QueryDataResponse, error) {
	var ts GrafanaSearchRequest

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	log.DefaultLogger.Debug("compartmentsResponse")
	log.DefaultLogger.Debug(ts.QueryType)
	log.DefaultLogger.Debug(ts.Region)
	log.DefaultLogger.Debug(ts.TenancyMode)
	log.DefaultLogger.Debug(ts.TenancyConfig)
	log.DefaultLogger.Debug(takey)

	var tenancyocid string
	if ts.TenancyConfig != "NoTenancyConfig" && ts.TenancyConfig != "" {
		res := strings.Split(takey, "/")
		tenancyocid = res[1]
	} else {
		tenancyocid = ts.TenancyOCID
	}

	log.DefaultLogger.Debug(tenancyocid)
	log.DefaultLogger.Debug("/compartmentsResponse")

	// if o.timeCacheUpdated.IsZero() || time.Now().Sub(o.timeCacheUpdated) > cacheRefreshTime {
	m, err := o.getCompartments(ctx, ts.Region, tenancyocid, takey)
	if err != nil {
		o.logger.Error("Unable to refresh cache")
		return nil, err
	}
	o.nameToOCID = m
	// }

	frame := data.NewFrame(query.RefID,
		data.NewField("name", nil, []string{}),
		data.NewField("compartmentID", nil, []string{}),
	)
	for name, id := range o.nameToOCID {
		frame.AppendRow(name, id)
	}

	return &backend.QueryDataResponse{
		Responses: map[string]backend.DataResponse{
			query.RefID: {
				Frames: data.Frames{frame},
			},
		},
	}, nil
}

func (o *OCIDatasource) getCompartments(ctx context.Context, region string, rootCompartment string, takey string) (map[string]string, error) {
	m := make(map[string]string)

	tenancyOcid := rootCompartment

	req := identity.GetTenancyRequest{TenancyId: common.String(tenancyOcid)}
	log.DefaultLogger.Debug(*req.TenancyId)

	// Send the request using the service client
	resp, err := o.tenancyAccess[takey].identityClient.GetTenancy(context.Background(), req)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("This is what we were trying to get %s", " : fetching tenancy name"))
	}

	mapFromIdToName := make(map[string]string)
	mapFromIdToName[tenancyOcid] = *resp.Name //tenancy name

	mapFromIdToParentCmptId := make(map[string]string)
	mapFromIdToParentCmptId[tenancyOcid] = "" //since root cmpt does not have a parent

	var page *string
	reg := common.StringToRegion(region)
	o.tenancyAccess[takey].identityClient.SetRegion(string(reg))
	for {
		res, err := o.tenancyAccess[takey].identityClient.ListCompartments(ctx,
			identity.ListCompartmentsRequest{
				CompartmentId:          &rootCompartment,
				Page:                   page,
				AccessLevel:            identity.ListCompartmentsAccessLevelAny,
				CompartmentIdInSubtree: common.Bool(true),
			})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("this is what we were trying to get %s", rootCompartment))
		}
		for _, compartment := range res.Items {
			if compartment.LifecycleState == identity.CompartmentLifecycleStateActive {
				mapFromIdToName[*(compartment.Id)] = *(compartment.Name)
				mapFromIdToParentCmptId[*(compartment.Id)] = *(compartment.CompartmentId)
			}
		}
		if res.OpcNextPage == nil {
			break
		}
		page = res.OpcNextPage
	}

	mapFromIdToFullCmptName := make(map[string]string)
	mapFromIdToFullCmptName[tenancyOcid] = mapFromIdToName[tenancyOcid] + "(tenancy, shown as '/')"

	for len(mapFromIdToFullCmptName) < len(mapFromIdToName) {
		for cmptId, cmptParentCmptId := range mapFromIdToParentCmptId {
			_, isCmptNameResolvedFullyAlready := mapFromIdToFullCmptName[cmptId]
			if !isCmptNameResolvedFullyAlready {
				if cmptParentCmptId == tenancyOcid {
					// If tenancy/rootCmpt my parent
					// cmpt name itself is fully qualified, just prepend '/' for tenancy aka rootCmpt
					mapFromIdToFullCmptName[cmptId] = "/" + mapFromIdToName[cmptId]
				} else {
					fullNameOfParentCmpt, isMyParentNameResolvedFully := mapFromIdToFullCmptName[cmptParentCmptId]
					if isMyParentNameResolvedFully {
						mapFromIdToFullCmptName[cmptId] = fullNameOfParentCmpt + "/" + mapFromIdToName[cmptId]
					}
				}
			}
		}
	}

	for cmptId, fullyQualifiedCmptName := range mapFromIdToFullCmptName {
		m[fullyQualifiedCmptName] = cmptId
	}

	return m, nil
}

type responseAndQuery struct {
	ociRes       monitoring.SummarizeMetricsDataResponse
	query        backend.DataQuery
	err          error
	legendFormat string
}

func (o *OCIDatasource) queryResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	results := make([]responseAndQuery, 0, len(req.Queries))
	var takey string

	for _, query := range req.Queries {
		var ts GrafanaOCIRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		fromMs := query.TimeRange.From.UnixNano() / int64(time.Millisecond)
		toMs := query.TimeRange.To.UnixNano() / int64(time.Millisecond)
		start := time.Unix(fromMs/1000, (fromMs%1000)*1000000).UTC()
		end := time.Unix(toMs/1000, (toMs%1000)*1000000).UTC()

		start = start.Truncate(time.Millisecond)
		end = end.Truncate(time.Millisecond)

		req := monitoring.SummarizeMetricsDataDetails{}
		req.Query = common.String(ts.Query)
		req.Namespace = common.String(ts.Namespace)
		req.Resolution = common.String(ts.Resolution)
		req.StartTime = &common.SDKTime{Time: start}
		req.EndTime = &common.SDKTime{Time: end}
		if ts.ResourceGroup != "NoResourceGroup" {
			req.ResourceGroup = common.String(ts.ResourceGroup)
		}

		// compute takey at every cycle of  queryResponse to guarantee mixed mode dashboards (single-multi or multi with different tenancies)
		if ts.TenancyConfig != "NoTenancyConfig" && ts.TenancyConfig != "" {
			takey = ts.TenancyConfig
		} else {
			takey = SingleTenancyKey
		}

		reg := common.StringToRegion(ts.Region)
		o.tenancyAccess[takey].metricsClient.SetRegion(string(reg))

		request := monitoring.SummarizeMetricsDataRequest{
			CompartmentId:               common.String(ts.Compartment),
			SummarizeMetricsDataDetails: req,
		}
		log.DefaultLogger.Debug("checkpoint 10")
		log.DefaultLogger.Debug(ts.Region)
		log.DefaultLogger.Debug(ts.Compartment)
		log.DefaultLogger.Debug(ts.QueryType)
		log.DefaultLogger.Debug(takey)
		log.DefaultLogger.Debug(ts.TenancyMode)
		log.DefaultLogger.Debug(ts.TenancyConfig)

		res, err := o.tenancyAccess[takey].metricsClient.SummarizeMetricsData(ctx, request)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint(spew.Sdump(query), spew.Sdump(request), spew.Sdump(res)))
		}

		log.DefaultLogger.Debug("checkpoint 20")

		// Include the legend format in the information about each query
		// since the legend format may be different for different queries
		// on the same data panel
		results = append(results, responseAndQuery{
			res,
			query,
			err,
			ts.LegendFormat,
		})
	}
	resp := backend.NewQueryDataResponse()
	for _, q := range results {
		respD := resp.Responses[q.query.RefID]

		if q.err != nil {
			respD.Error = fmt.Errorf(q.err.Error())
			continue
		}

		for _, item := range q.ociRes.Items {
			metricName := *(item.Name)

			// NOTE: There are a few OCI resources, e.g. SCH, for which no such
			// dimension is defined!!!
			if resourceIdValue, ok := item.Dimensions["resourceId"]; ok {
				item.Dimensions["resourceId"] = strings.ToLower(resourceIdValue)
			} else if resourceIdValue, ok := item.Dimensions["resourceID"]; ok {
				item.Dimensions["resourceID"] = strings.ToLower(resourceIdValue)
			}

			dimensionKeys := make([]string, len(item.Dimensions))
			i := 0

			for key := range item.Dimensions {
				dimensionKeys[i] = key
				i++
			}

			var fullDisplayName string
			// If the legend format field in the query editor is empty then the metric label will be:
			//   <Metric name>[<dimension value 1> | <dimension value 2> | ... <dimension value N>]
			if q.legendFormat == "" {
				sort.Strings(dimensionKeys)

				var dmValueListForMetricStream = ""
				for _, dimensionKey := range dimensionKeys {
					var dimValue = item.Dimensions[dimensionKey]

					if dmValueListForMetricStream == "" {
						dmValueListForMetricStream = "[" + dimValue
					} else {
						dmValueListForMetricStream = dmValueListForMetricStream + " | " + dimValue
					}

				}
				dmValueListForMetricStream = dmValueListForMetricStream + "]"
				fullDisplayName = metricName + dmValueListForMetricStream
				// If user has provided a value for the legend format then use the format to
				// generate the display name for the metric
			} else {
				fullDisplayName = o.generateCustomMetricLabel(q.legendFormat, metricName, item.Dimensions)
			}

			//dimeString, _ := json.Marshal(item.Dimensions)
			var fieldConfig = data.FieldConfig{}

			if _, okMinRange := item.Metadata["minRange"]; okMinRange {
				minFloat, err := strconv.ParseFloat(item.Metadata["minRange"], 64)
				if err == nil {
					fieldConfig = *(&fieldConfig).SetMin(minFloat)
				}
			}
			if _, okMinRange := item.Metadata["maxRange"]; okMinRange {
				maxFloat, err := strconv.ParseFloat(item.Metadata["maxRange"], 64)
				if err == nil {
					fieldConfig = *(&fieldConfig).SetMax(maxFloat)
				}
			}

			if _, okUnitName := item.Metadata["unit"]; okUnitName {
				fieldConfig.Unit = item.Metadata["unit"]
			}

			fieldConfig.DisplayNameFromDS = fullDisplayName

			frame := data.NewFrame(q.query.RefID,
				data.NewField("Time", nil, []time.Time{}),
				data.NewField("Value", item.Dimensions, []float64{}).SetConfig(&fieldConfig),
			)

			for _, metric := range item.AggregatedDatapoints {
				frame.AppendRow(metric.Timestamp.Time, *(metric.Value))
			}

			respD.Frames = append(respD.Frames, frame)
			resp.Responses[q.query.RefID] = respD
		}
	}
	return resp, nil
}

func (o *OCIDatasource) regionsResponse(ctx context.Context, req *backend.QueryDataRequest, takey string) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaOCIRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		res, err := o.tenancyAccess[takey].identityClient.ListRegions(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching regions")
		}

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))
		var regionName []string

		/* Generate list of regions */
		for _, item := range res.Items {
			regionName = append(regionName, *(item.Name))
		}

		/* Sort regions list */
		sort.Strings(regionName)
		for _, sortedRegions := range regionName {
			frame.AppendRow(sortedRegions)
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
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
	mDimensions map[string]string) string {

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
				// Check whether there is a dimension name for the metric that matches
				// the placeholder label. If there is then replace the placeholder with
				// the value of the dimension
				if dimensionValue, ok := mDimensions[placeholderLabel]; ok {
					metricLabel = re.ReplaceAllString(metricLabel, dimensionValue)
				}
			}
		}
	}
	o.logger.Debug("Generated metric label", "legendFormat", legendFormat,
		"metricName", metricName, "metricLabel", metricLabel)
	return metricLabel
}

/*
Function generates an array  containing OCI configuration (.oci/config) in the following format:

<section label/TenancyOCID>

*/

func (o *OCIDatasource) tenancyConfigResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()
	ociconfigs, _ := OCIConfigParser("ociconfigs")

	for _, query := range req.Queries {
		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))
		for _, ociconfig := range ociconfigs {
			configProvider := common.CustomProfileConfigProvider("", ociconfig)
			res, err := configProvider.TenancyOCID()
			if err != nil {
				return nil, errors.Wrap(err, "error configuring TenancyOCID: "+ociconfig+"/"+res)
			}
			value := ociconfig + "/" + res
			frame.AppendRow(*(common.String(value)))
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
}

/*
Function parses the content of .oci/config file
It accepts one parameter (scope) which can be "ociconfigs" or "regions"

if ociconfigs, then the function returns an array of the OCI config sections labels
if regions, then the function returns the list of the regions of every OCI config section entries
*/
func OCIConfigParser(scope string) ([]string, error) {
	var oci_config_file string
	var ociconfigs []string
	var regions []string

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.DefaultLogger.Error("could not get home directory")
	}

	if _, ok := os.LookupEnv("OCI_CLI_CONFIG_FILE"); ok {
		oci_config_file = os.Getenv("OCI_CLI_CONFIG_FILE")
	} else {
		oci_config_file = homedir + "/.oci/config"
	}

	file, err := os.Open(oci_config_file)
	if err != nil {
		return nil, errors.Wrap(err, "error opening file:"+oci_config_file)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "buffer error")
	}

	r, err := regexp.Compile(`\[.*\]`) // this can also be a regex
	if err != nil {
		return nil, errors.Wrap(err, "error in compiling regex")
	}
	r2, err := regexp.Compile(`region=`) // this can also be a regex
	if err != nil {
		return nil, errors.Wrap(err, "error in compiling regex")
	}

	if scope == "ociconfigs" {
		for scanner.Scan() {
			if r.MatchString(scanner.Text()) {
				replacer := strings.NewReplacer("[", "", "]", "")
				output := replacer.Replace(scanner.Text())
				ociconfigs = append(ociconfigs, output)
			}
		}
		return ociconfigs, nil
	}

	if scope == "regions" {
		for scanner.Scan() {
			if r2.MatchString(scanner.Text()) {
				replacer := strings.NewReplacer("region=", "")
				output := replacer.Replace(scanner.Text())
				regions = append(regions, output)

			}
		}
		return regions, nil
	}
	return nil, nil
}
