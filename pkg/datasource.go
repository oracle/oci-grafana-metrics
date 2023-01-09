// Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
const NoTenancy = "NoTenancy"

var profileRegex = regexp.MustCompile(`^\[(.*)\]`)

var (
	cacheRefreshTime = time.Minute // how often to refresh our compartmentID cache
	re               = regexp.MustCompile(`(?m)\w+Name`)
)

type OCIDatasource struct {
	tenancyAccess    map[string]*TenancyAccess
	logger           log.Logger
	nameToOCID       map[string]string
	timeCacheUpdated time.Time
}

// NewOCIConfigFile - constructor
func NewOCIConfigFile() *OCIConfigFile {
	return &OCIConfigFile{
		tenancyocid: make(map[string]string),
		region:      make(map[string]string),
		user:        make(map[string]string),
		logger:      log.DefaultLogger,
	}
}

// NewOCIDatasource - constructor
func NewOCIDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &OCIDatasource{
		tenancyAccess: make(map[string]*TenancyAccess),
		logger:        log.DefaultLogger,
		nameToOCID:    make(map[string]string),
	}, nil
}

type OCIConfigFile struct {
	tenancyocid map[string]string
	region      map[string]string
	user        map[string]string
	logger      log.Logger
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
	ResourceGroup string
	LegendFormat  string
}

// GrafanaSearchRequest incoming request body for search requests
type GrafanaSearchRequest struct {
	GrafanaCommonRequest
	Metric        string `json:"metric,omitempty"`
	Namespace     string
	ResourceGroup string
}

// GrafanaCommonRequest - captures the common parts of the search and metricsRequests
type GrafanaCommonRequest struct {
	Compartment string
	Environment string
	TenancyMode string
	QueryType   string
	Region      string
	Tenancy     string // the actual tenancy with the format <configfile entry name/tenancyOCID>
	TenancyOCID string `json:"tenancyOCID"`
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

	if len(o.tenancyAccess) == 0 || ts.TenancyMode == "multitenancy" {
		err := o.getConfigProvider(ts.Environment, ts.TenancyMode)
		if err != nil {
			return nil, errors.Wrap(err, "broken environment")
		}
	}

	if ts.TenancyMode == "multitenancy" {
		takey = ts.Tenancy
	} else {
		takey = SingleTenancyKey
	}

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
	case "tenancies":
		return o.tenanciesResponse(ctx, req, ts.Environment)
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

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	reg := common.StringToRegion(ts.Region)

	for key, _ := range o.tenancyAccess {
		if ts.TenancyMode == "multitenancy" {
			var p *OCIConfigFile
			var ociparsErr error
			var tenancyErr error
			if ts.Environment == "local" {
				oci_config_file := OCIConfigPath()
				p, ociparsErr = OCIConfigParser(oci_config_file)
				if ociparsErr != nil {
					return &backend.QueryDataResponse{}, errors.Wrap(ociparsErr, fmt.Sprintf("OCI Config Parser failed"))
				}
			} else {
				return &backend.QueryDataResponse{}, errors.Wrap(ociparsErr, fmt.Sprintf("Multitenancy mode using instance principals is not implemented yet."))
			}
			res := strings.Split(key, "/")
			tenancyocid, tenancyErr = o.tenancyAccess[key].config.TenancyOCID()
			if tenancyErr != nil {
				return nil, errors.Wrap(tenancyErr, "error fetching TenancyOCID")
			}
			reg = common.StringToRegion(p.region[res[0]])
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
			o.logger.Debug(key, "OK", status)
		} else {
			o.logger.Debug(key, "FAILED", status)
			return nil, errors.Wrap(err, fmt.Sprintf("list metrics failed %s %d", spew.Sdump(res), status))
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
	var p *OCIConfigFile
	var ociparsErr error

	switch environment {
	case "local":
		oci_config_file := OCIConfigPath()
		if tenancymode == "multitenancy" {
			p, ociparsErr = OCIConfigParser(oci_config_file)
			if ociparsErr != nil {
				return errors.Wrap(ociparsErr, fmt.Sprintf("OCI Config Parser failed"))
			}
			for key, _ := range p.tenancyocid {
				var configProvider common.ConfigurationProvider
				configProvider = common.CustomProfileConfigProvider(oci_config_file, key)
				metricsClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
				if err != nil {
					o.logger.Error("Error with config:" + key)
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
				o.tenancyAccess[key+"/"+tenancyocid] = &TenancyAccess{metricsClient, identityClient, configProvider}
			}
			return nil
		} else {
			var configProvider common.ConfigurationProvider
			configProvider = common.CustomProfileConfigProvider(oci_config_file, "DEFAULT")
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

	var tenancyocid string
	if ts.TenancyMode == "multitenancy" {
		if len(takey) <= 0 || takey == NoTenancy {
			o.logger.Error("Unable to get Multi-tenancy OCID")
			err := fmt.Errorf("Tenancy OCID %s is not valid.", takey)
			return nil, err
		}
		res := strings.Split(takey, "/")
		tenancyocid = res[1]
	} else {
		tenancyocid = ts.TenancyOCID
	}

	if o.timeCacheUpdated.IsZero() || time.Now().Sub(o.timeCacheUpdated) > cacheRefreshTime {
		m, err := o.getCompartments(ctx, ts.Region, tenancyocid, takey)
		if err != nil {
			o.logger.Error("Unable to refresh cache")
			return nil, err
		}
		o.nameToOCID = m
	}

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

	reg := common.StringToRegion(region)
	o.tenancyAccess[takey].identityClient.SetRegion(string(reg))

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

		// compute takey at every cycle of queryResponse to guarantee mixed mode dashboards (single-multi or multi with different tenancies)
		if ts.TenancyMode == "multitenancy" {
			takey = ts.Tenancy
			if len(takey) <= 0 || takey == NoTenancy {
				o.logger.Error("Unable to get Multi-tenancy OCID")
				err := fmt.Errorf("Tenancy OCID %s is not valid.", takey)
				return nil, err
			}
		} else {
			takey = SingleTenancyKey
		}

		reg := common.StringToRegion(ts.Region)
		o.tenancyAccess[takey].metricsClient.SetRegion(string(reg))

		request := monitoring.SummarizeMetricsDataRequest{
			CompartmentId:               common.String(ts.Compartment),
			SummarizeMetricsDataDetails: req,
		}

		res, err := o.tenancyAccess[takey].metricsClient.SummarizeMetricsData(ctx, request)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint(spew.Sdump(query), spew.Sdump(request), spew.Sdump(res)))
		}

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
func (o *OCIDatasource) tenanciesResponse(ctx context.Context, req *backend.QueryDataRequest, env string) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()
	var p *OCIConfigFile
	var res string
	oci_config_file := OCIConfigPath()
	p, err := OCIConfigParser(oci_config_file)
	if err != nil {
		log.DefaultLogger.Error("could not parse config file")
		return nil, err
	}
	for _, query := range req.Queries {
		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))
		// for _, ociconfig := range ociconfigs {
		for key, _ := range p.tenancyocid {
			if env == "local" {
				res = p.tenancyocid[key]
			} else {
				configProvider := common.CustomProfileConfigProvider(oci_config_file, key)
				res, err := configProvider.TenancyOCID()
				if err != nil {
					return nil, errors.Wrap(err, "error configuring TenancyOCID: "+key+"/"+res)
				}
			}
			value := key + "/" + res
			frame.AppendRow(*(common.String(value)))
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
}

/*
Function parses the content of .oci/config file and returns raw file content.
It then pass over to parseConfigFile in search for each config entry.
*/
func OCIConfigParser(oci_config_file string) (*OCIConfigFile, error) {
	p := NewOCIConfigFile()
	data, err := ioutil.ReadFile(oci_config_file)
	if err != nil {
		err = fmt.Errorf("can not read config file: %s due to: %s", oci_config_file, err.Error())
		return nil, err
	}
	if len(data) == 0 {
		err = fmt.Errorf("config file %s is empty.", oci_config_file)
		return nil, err
	}
	err = p.parseConfigFile(data)
	if err != nil {
		log.DefaultLogger.Error("config file " + oci_config_file + " is not valid.")
		return nil, err
	}
	return p, nil
}

/*
Function parses the content of .oci/config file
It looks for each profile entry and pass over to the parseConfigAtLine function
*/
func (p *OCIConfigFile) parseConfigFile(data []byte) (err error) {
	content := string(data)
	splitContent := strings.Split(content, "\n")
	if len(splitContent) == 0 {
		err = fmt.Errorf("config file is corrupted.")
		return err
	}
	//Look for profile
	for i, line := range splitContent {
		if match := profileRegex.FindStringSubmatch(line); match != nil && len(match) > 1 {
			start := i + 1
			p.parseConfigAtLine(start, match[1], splitContent)
		}
	}
	if len(p.tenancyocid) == 0 {
		err = fmt.Errorf("config file is not valid.")
		return err
	}
	return nil
}

/*
Function parses the output of parseConfigFile function looking for specific entries.
user, tenancy and region are retrieved and stored in the OCIConfigFile maps.
*/
func (p *OCIConfigFile) parseConfigAtLine(start int, profile string, content []string) (err error) {
	for i := start; i < len(content); i++ {
		line := content[i]
		if profileRegex.MatchString(line) {
			break
		}
		if !strings.Contains(line, "=") {
			continue
		}
		splits := strings.Split(line, "=")
		switch key, value := strings.TrimSpace(splits[0]), strings.TrimSpace(splits[1]); strings.ToLower(key) {
		case "user":
			p.user[profile] = value
		case "tenancy":
			p.tenancyocid[profile] = value
		case "region":
			p.region[profile] = value
		}
	}
	return
}

/*
Function returns the path for the .oci/config file
*/
func OCIConfigPath() string {
	var oci_config_file string
	homedir := "/usr/share/grafana"
	if _, ok := os.LookupEnv("OCI_CLI_CONFIG_FILE"); ok {
		oci_config_file = os.Getenv("OCI_CLI_CONFIG_FILE")
	} else {
		oci_config_file = homedir + "/.oci/config"
	}
	return oci_config_file
}
