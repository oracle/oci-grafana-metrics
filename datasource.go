// Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/oracle/oci-go-sdk/monitoring"
	"github.com/pkg/errors"
)

const MaxPagesToFetch = 20

var (
	cacheRefreshTime = time.Minute // how often to refresh our compartmentID cache
	re               = regexp.MustCompile(`(?m)\w+Name`)
)

//OCIDatasource - pulls in data from telemtry/various oci apis
type OCIDatasource struct {
	metricsClient    monitoring.MonitoringClient
	identityClient   identity.IdentityClient
	config           common.ConfigurationProvider
	cmptid           string
	logger           log.Logger
	nameToOCID       map[string]string
	timeCacheUpdated time.Time
}

//NewOCIDatasource - constructor
func NewOCIDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &OCIDatasource{
		logger:     log.DefaultLogger,
		nameToOCID: make(map[string]string),
	}, nil
}

// GrafanaOCIRequest - Query Request comning in from the front end
type GrafanaOCIRequest struct {
	GrafanaCommonRequest
	Query         string
	Resolution    string
	Namespace     string
	ResourceGroup string
}

//GrafanaSearchRequest incoming request body for search requests
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
	QueryType   string
	Region      string
	TenancyOCID string `json:"tenancyOCID"`
}

// Query - Determine what kind of query we're making
func (o *OCIDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	var ts GrafanaCommonRequest

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	queryType := ts.QueryType
	if o.config == nil {
		configProvider, err := getConfigProvider(ts.Environment)
		if err != nil {
			return nil, errors.Wrap(err, "broken environment")
		}
		metricsClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
		if err != nil {
			return nil, errors.New(fmt.Sprint("error with client", spew.Sdump(configProvider), err.Error()))
		}
		identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
		if err != nil {
			o.logger.Error("error with client")
			panic(err)
		}
		o.identityClient = identityClient
		o.metricsClient = metricsClient
		o.config = configProvider
		if ts.Compartment != "" {
			o.cmptid = ts.Compartment
		}
	}

	switch queryType {
	case "compartments":
		return o.compartmentsResponse(ctx, req)
	case "dimensions":
		return o.dimensionResponse(ctx, req)
	case "namespaces":
		return o.namespaceResponse(ctx, req)
	case "resourcegroups":
		return o.resourcegroupsResponse(ctx, req)
	case "regions":
		return o.regionsResponse(ctx, req)
	case "search":
		return o.searchResponse(ctx, req)
	case "test":
		return o.testResponse(ctx, req)
	default:
		return o.queryResponse(ctx, req)
	}
}

func (o *OCIDatasource) testResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	var ts GrafanaCommonRequest

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	//o.logger.Error("ts.Com is " + ts.Compartment)
	listMetrics := monitoring.ListMetricsRequest{
		CompartmentId: common.String(ts.Compartment),
	}
	reg := common.StringToRegion(ts.Region)
	o.metricsClient.SetRegion(string(reg))
	res, err := o.metricsClient.ListMetrics(ctx, listMetrics)
	if err != nil {
		return &backend.QueryDataResponse{}, err
	}
	status := res.RawResponse.StatusCode
	if status >= 200 && status < 300 {
		return &backend.QueryDataResponse{}, nil
	}
	return nil, errors.Wrap(err, fmt.Sprintf("list metrircs failed %s %d", spew.Sdump(res), status))
}

func (o *OCIDatasource) dimensionResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
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
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails)
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

func (o *OCIDatasource) namespaceResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaSearchRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		reqDetails := monitoring.ListMetricsDetails{}
		reqDetails.GroupBy = []string{"namespace"}
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails)
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

func (o *OCIDatasource) resourcegroupsResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaSearchRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}

		reqDetails := monitoring.ListMetricsDetails{}
		reqDetails.Namespace = common.String(ts.Namespace)
		reqDetails.GroupBy = []string{"resourceGroup"}
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails)
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

func getConfigProvider(environment string) (common.ConfigurationProvider, error) {
	switch environment {
	case "local":
		return common.DefaultConfigProvider(), nil
	case "OCI Instance":
		return auth.InstancePrincipalConfigurationProvider()
	default:
		return nil, errors.New("unknown environment type")
	}
}

func (o *OCIDatasource) searchResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
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

		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails)
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

func (o *OCIDatasource) searchHelper(ctx context.Context, region, compartment string, metricDetails monitoring.ListMetricsDetails) ([]monitoring.Metric, error) {
	var items []monitoring.Metric
	var page *string

	pageNumber := 0
	for {
		reg := common.StringToRegion(region)
		o.metricsClient.SetRegion(string(reg))
		res, err := o.metricsClient.ListMetrics(ctx, monitoring.ListMetricsRequest{
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

func (o *OCIDatasource) compartmentsResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	var ts GrafanaSearchRequest

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	if o.timeCacheUpdated.IsZero() || time.Now().Sub(o.timeCacheUpdated) > cacheRefreshTime {
		m, err := o.getCompartments(ctx, ts.Region, ts.TenancyOCID)
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

func (o *OCIDatasource) getCompartments(ctx context.Context, region string, rootCompartment string) (map[string]string, error) {
	m := make(map[string]string)

	tenancyOcid := rootCompartment

	req := identity.GetTenancyRequest{TenancyId: common.String(tenancyOcid)}
	// Send the request using the service client
	resp, err := o.identityClient.GetTenancy(context.Background(), req)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("This is what we were trying to get %s", " : fetching tenancy name"))
	}

	mapFromIdToName := make(map[string]string)
	mapFromIdToName[tenancyOcid] = *resp.Name //tenancy name

	mapFromIdToParentCmptId := make(map[string]string)
	mapFromIdToParentCmptId[tenancyOcid] = "" //since root cmpt does not have a parent

	var page *string
	reg := common.StringToRegion(region)
	o.identityClient.SetRegion(string(reg))
	for {
		res, err := o.identityClient.ListCompartments(ctx,
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
		if o.cmptid == cmptId || o.isCmptMetricsAllowed(ctx, region, cmptId, fullyQualifiedCmptName) {
			m[fullyQualifiedCmptName] = cmptId
		}
	}
	return m, nil
}

func (o *OCIDatasource) isCmptMetricsAllowed(ctx context.Context, region string, cmptId string, cmptName string) bool {
	reg := common.StringToRegion(region)
	o.metricsClient.SetRegion(string(reg))

	res, _ := o.metricsClient.ListMetrics(ctx, monitoring.ListMetricsRequest{
		CompartmentId: common.String(cmptId),
		Page:          nil,
	})

	if res.RawResponse.StatusCode == 200 { // access allowed
		o.logger.Error("Success Access to cmptId:" + cmptId + "with name:" + cmptName)
		return true
	} else if res.RawResponse.StatusCode == 429 { // throttled? try again
		time.Sleep(4 * time.Second)
		res1, _ := o.metricsClient.ListMetrics(ctx, monitoring.ListMetricsRequest{
			CompartmentId: common.String(cmptId),
			Page:          nil,
		})
		if res1.RawResponse.StatusCode == 200 { // access allowed
			return true
		} else {
			o.logger.Error("error loading filtered compartments")
			return false
		}
	} else if res.RawResponse.StatusCode == 404 { // access allowed
		//o.logger.Error("404 Access to cmptId:" + cmptId + "with name:" + cmptName)
	} else if res.RawResponse.StatusCode == 401 { // access allowed
		//o.logger.Error("401 Access to cmptId:" + cmptId + "with name:" + cmptName)
	}
	return false
}

type responseAndQuery struct {
	ociRes monitoring.SummarizeMetricsDataResponse
	query  backend.DataQuery
	err    error
}

func (o *OCIDatasource) queryResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	results := make([]responseAndQuery, 0, len(req.Queries))

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

		reg := common.StringToRegion(ts.Region)
		o.metricsClient.SetRegion(string(reg))

		request := monitoring.SummarizeMetricsDataRequest{
			CompartmentId:               common.String(ts.Compartment),
			SummarizeMetricsDataDetails: req,
		}

		res, err := o.metricsClient.SummarizeMetricsData(ctx, request)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint(spew.Sdump(query), spew.Sdump(request), spew.Sdump(res)))
		}
		results = append(results, responseAndQuery{
			res,
			query,
			err,
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
			name := *(item.Name)

			item.Dimensions["resourceId"] = strings.ToLower(item.Dimensions["resourceId"])
			dimensionKeys := make([]string, len(item.Dimensions))
			i := 0

			for key := range item.Dimensions {
				dimensionKeys[i] = key
				i++
			}

			sort.Strings(dimensionKeys)

			var dmValueListForMetricStream = ""
			for _, dimensionKey := range dimensionKeys {
				var dimValue = item.Dimensions[dimensionKey]

				if strings.HasPrefix(dimValue, "ocid1.") {
					ocidParts := strings.SplitAfter(dimValue, ".")
					if len(ocidParts[len(ocidParts)-1]) > 5 {
						var alphaOcidId = ocidParts[len(ocidParts)-1]
						ocidParts[len(ocidParts)-1] = "XXXX" + alphaOcidId[len(alphaOcidId)-6:]
					}
					dimValue = strings.Join(ocidParts, "")
				}

				if dmValueListForMetricStream == "" {
					dmValueListForMetricStream = "[" + dimValue
				} else {
					dmValueListForMetricStream = dmValueListForMetricStream + "," + dimValue
				}
			}
			dmValueListForMetricStream = dmValueListForMetricStream + "]"
			fullDisplayName := name + dmValueListForMetricStream

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

			o.logger.Info("item.Metadata " + fmt.Sprint(item.Metadata))

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

func (o *OCIDatasource) regionsResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var ts GrafanaOCIRequest
		if err := json.Unmarshal(query.JSON, &ts); err != nil {
			return &backend.QueryDataResponse{}, err
		}
		res, err := o.identityClient.ListRegions(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching regions")
		}

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))

		for _, item := range res.Items {
			frame.AppendRow(*(item.Name))
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
}
