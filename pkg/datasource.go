// Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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

var EmptyString string = ""
var EmptyKeyPass *string = &EmptyString

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
		fingerprint: make(map[string]string),
		privkey:     make(map[string]string),
		privkeypass: make(map[string]*string),
		logger:      log.DefaultLogger,
	}
}

// NewOCIDatasourceConstructor - constructor
func NewOCIDatasourceConstructor() *OCIDatasource {
	return &OCIDatasource{
		tenancyAccess: make(map[string]*TenancyAccess),
		logger:        log.DefaultLogger,
		nameToOCID:    make(map[string]string),
	}
}

func NewOCIDatasource(req backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var ts GrafanaCommonRequest
	log.DefaultLogger.Error("NewOCIDatasource")

	o := NewOCIDatasourceConstructor()

	if err := json.Unmarshal(req.JSONData, &ts); err != nil {
		return nil, fmt.Errorf("can not read settings: %s", err.Error())
	}

	o.logger.Debug("check1 " + ts.Environment)
	o.logger.Debug("check1 " + ts.TenancyMode)

	if len(o.tenancyAccess) == 0 {
		err := o.getConfigProvider(ts.Environment, ts.TenancyMode, req)
		if err != nil {
			return nil, errors.Wrap(err, "broken environment")
		}
	}

	o.logger.Debug(ts.Environment)
	o.logger.Debug(ts.TenancyMode)
	if len(o.tenancyAccess) == 0 {
		o.logger.Debug("vuoto")
	} else {
		o.logger.Debug("Pieno")
	}

	// return &OCIDatasource{
	// 	tenancyAccess: make(map[string]*TenancyAccess),
	// 	logger:        log.DefaultLogger,
	// 	nameToOCID:    make(map[string]string),
	// }, nil
	return &o, nil
}

type OCIConfigFile struct {
	tenancyocid map[string]string
	region      map[string]string
	user        map[string]string
	fingerprint map[string]string
	privkey     map[string]string
	privkeypass map[string]*string
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

type OCISecuredSettings struct {
	Profile_0     string `json:"profile0,omitempty"`
	Tenancy_0     string `json:"tenancy0,omitempty"`
	Region_0      string `json:"region0,omitempty"`
	User_0        string `json:"user0,omitempty"`
	Privkey_0     string `json:"privkey0,omitempty"`
	Fingerprint_0 string `json:"fingerprint0,omitempty"`

	Profile_1     string `json:"profile1,omitempty"`
	Tenancy_1     string `json:"tenancy1,omitempty"`
	Region_1      string `json:"region1,omitempty"`
	User_1        string `json:"user1,omitempty"`
	Fingerprint_1 string `json:"fingerprint1,omitempty"`
	Privkey_1     string `json:"privkey1,omitempty"`

	Profile_2     string `json:"profile2,omitempty"`
	Tenancy_2     string `json:"tenancy2,omitempty"`
	Region_2      string `json:"region2,omitempty"`
	User_2        string `json:"user2,omitempty"`
	Fingerprint_2 string `json:"fingerprint2,omitempty"`
	Privkey_2     string `json:"privkey2,omitempty"`

	Profile_3     string `json:"profile3,omitempty"`
	Tenancy_3     string `json:"tenancy3,omitempty"`
	Region_3      string `json:"region3,omitempty"`
	User_3        string `json:"user3,omitempty"`
	Fingerprint_3 string `json:"fingerprint3,omitempty"`
	Privkey_3     string `json:"privkey3,omitempty"`

	Profile_4     string `json:"profile4,omitempty"`
	Tenancy_4     string `json:"tenancy4,omitempty"`
	Region_4      string `json:"region4,omitempty"`
	User_4        string `json:"user4,omitempty"`
	Fingerprint_4 string `json:"fingerprint4,omitempty"`
	Privkey_4     string `json:"privkey4,omitempty"`

	Profile_5     string `json:"profile5,omitempty"`
	Tenancy_5     string `json:"tenancy5,omitempty"`
	Region_5      string `json:"region5,omitempty"`
	User_5        string `json:"user5,omitempty"`
	Fingerprint_5 string `json:"fingerprint5,omitempty"`
	Privkey_5     string `json:"privkey5,omitempty"`
}

// Prepare format to decode SecureJson
func transcode(in, out interface{}) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(in)
	json.NewDecoder(buf).Decode(out)
}

// Query - Determine what kind of query we're making
func (o *OCIDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Error("QueryData Checkpoin 0")
	var ts GrafanaCommonRequest
	var takey string

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	queryType := ts.QueryType

	o.logger.Debug("QueryType")
	o.logger.Debug(ts.QueryType)
	o.logger.Debug(ts.Environment)
	o.logger.Debug(ts.Tenancy)
	o.logger.Debug(ts.TenancyMode)

	// // if len(o.tenancyAccess) == 0 || ts.TenancyMode == "multitenancy" {
	// if len(o.tenancyAccess) == 0 {
	// 	err := o.getConfigProvider(ts.Environment, ts.TenancyMode, req)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "broken environment")
	// 	}
	// }

	if ts.TenancyMode == "multitenancy" {
		takey = ts.Tenancy
	} else {
		takey = SingleTenancyKey
	}
	o.logger.Debug(takey)
	o.logger.Debug("/QueryType")

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
		return o.tenanciesResponse(ctx, req)
	case "search":
		return o.searchResponse(ctx, req, takey)
	case "test":
		return o.testResponse(ctx, req)
	default:
		return o.queryResponse(ctx, req)
	}
}

func (o *OCIDatasource) getConfigProvider(environment string, tenancymode string, req backend.DataSourceInstanceSettings) error {
	switch environment {
	case "local":
		q, err := OCILoadSettings(req)
		if err != nil {
			o.logger.Error("Error Loading config settings", "error", err)
			return errors.Wrap(err, "Error Loading config settings")
		}
		for key, _ := range q.tenancyocid {
			var configProvider common.ConfigurationProvider
			configProvider = common.NewRawConfigurationProvider(q.tenancyocid[key], q.user[key], q.region[key], q.fingerprint[key], q.privkey[key], q.privkeypass[key])
			// configProvider = common.CustomProfileConfigProvider(oci_config_file, key)
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
			if tenancymode == "multitenancy" {
				o.tenancyAccess[key+"/"+tenancyocid] = &TenancyAccess{metricsClient, identityClient, configProvider}
			} else {
				o.tenancyAccess[SingleTenancyKey] = &TenancyAccess{metricsClient, identityClient, configProvider}
			}
		}
		o.logger.Debug("checkpint 1 getConfigProvider " + environment)
		return nil

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

func (o *OCIDatasource) testResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	var ts GrafanaCommonRequest
	var reg common.Region
	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}

	for key, _ := range o.tenancyAccess {
		if ts.TenancyMode == "multitenancy" && ts.Environment != "local" {
			var ociparsErr error
			return &backend.QueryDataResponse{}, errors.Wrap(ociparsErr, fmt.Sprintf("Multitenancy mode using instance principals is not implemented yet."))
		}
		tenancyocid, tenancyErr := o.tenancyAccess[key].config.TenancyOCID()
		if tenancyErr != nil {
			return nil, errors.Wrap(tenancyErr, "error fetching TenancyOCID")
		}
		regio, regErr := o.tenancyAccess[key].config.Region()
		if regErr != nil {
			return nil, errors.Wrap(regErr, "error fetching TenancyOCID")
		}
		reg = common.StringToRegion(regio)

		listMetrics := monitoring.ListMetricsRequest{
			CompartmentId: common.String(tenancyocid),
		}
		o.tenancyAccess[key].metricsClient.SetRegion(string(reg))
		res, err := o.tenancyAccess[key].metricsClient.ListMetrics(ctx, listMetrics)
		if err != nil {
			o.logger.Debug(key, "FAILED", err)
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
	var tenancyErr error

	if ts.TenancyMode == "multitenancy" {
		if len(takey) <= 0 || takey == NoTenancy {
			o.logger.Error("Unable to get Multi-tenancy OCID")
			err := fmt.Errorf("Tenancy OCID %s is not valid.", takey)
			return nil, err
		}
		res := strings.Split(takey, "/")
		tenancyocid = res[1]
	} else {
		tenancyocid, tenancyErr = o.tenancyAccess[takey].config.TenancyOCID()
		if tenancyErr != nil {
			return nil, errors.Wrap(tenancyErr, "error fetching TenancyOCID")
		}
	}

	regio, regErr := o.tenancyAccess[takey].config.Region()
	if regErr != nil {
		return nil, errors.Wrap(regErr, "error fetching TenancyOCID")
	}

	if o.timeCacheUpdated.IsZero() || time.Now().Sub(o.timeCacheUpdated) > cacheRefreshTime {
		m, err := o.getCompartments(ctx, tenancyocid, regio, takey)
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

func (o *OCIDatasource) getCompartments(ctx context.Context, rootCompartment string, region string, takey string) (map[string]string, error) {
	m := make(map[string]string)

	tenancyOcid := rootCompartment

	reg := common.StringToRegion(region)
	o.tenancyAccess[takey].metricsClient.SetRegion(string(reg))

	req := identity.GetTenancyRequest{TenancyId: common.String(tenancyOcid)}

	log.DefaultLogger.Error("getCompartments tenancyocid " + tenancyOcid)

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
		tenancyocid, tenancyErr := o.tenancyAccess[takey].config.TenancyOCID()
		if tenancyErr != nil {
			return nil, errors.Wrap(tenancyErr, "error fetching TenancyOCID")
		}
		req := identity.ListRegionSubscriptionsRequest{TenancyId: common.String(tenancyocid)}

		// Send the request using the service client
		res, err := o.tenancyAccess[takey].identityClient.ListRegionSubscriptions(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching regions")
		}

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))
		var regionName []string

		/* Generate list of regions */
		for _, item := range res.Items {
			regionName = append(regionName, *(item.RegionName))
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
func (o *OCIDatasource) tenanciesResponse(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	resp := backend.NewQueryDataResponse()
	for _, query := range req.Queries {

		frame := data.NewFrame(query.RefID, data.NewField("text", nil, []string{}))
		for key, _ := range o.tenancyAccess {
			frame.AppendRow(*(common.String(key)))
		}

		respD := resp.Responses[query.RefID]
		respD.Frames = append(respD.Frames, frame)
		resp.Responses[query.RefID] = respD
	}
	return resp, nil
}

// OCILoadSettings will read and validate Settings from the DataSourceConfig
func OCILoadSettings(req backend.DataSourceInstanceSettings) (*OCIConfigFile, error) {
	q := NewOCIConfigFile()

	TenancySettingsBlock := 0
	var dat OCISecuredSettings

	if err := json.Unmarshal(req.JSONData, &dat); err != nil {
		return nil, fmt.Errorf("can not read settings: %s", err.Error())
	}

	// password, ok := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"]
	// if ok {
	// 	dat.Fingerprint_0 = password
	// }
	decryptedJSONData := req.DecryptedSecureJSONData
	transcode(decryptedJSONData, &dat)

	v := reflect.ValueOf(dat)
	typeOfS := v.Type()
	var key string

	for FieldIndex := 0; FieldIndex < v.NumField(); FieldIndex++ {
		splits := strings.Split(typeOfS.Field(FieldIndex).Name, "_")
		SettingsBlockIndex, interr := strconv.Atoi(splits[1])
		if interr != nil {
			return nil, fmt.Errorf("can not read settings: %s", interr.Error())
		}
		if SettingsBlockIndex == TenancySettingsBlock {
			if splits[0] == "Profile" {
				if v.Field(FieldIndex).Interface() != "" {
					key = fmt.Sprintf("%v", v.Field(FieldIndex).Interface())
				} else {
					return q, nil
				}
			} else {
				log.DefaultLogger.Error(key)
				log.DefaultLogger.Error(splits[0])

				switch value := v.Field(FieldIndex).Interface(); strings.ToLower(splits[0]) {
				case "tenancy":
					q.tenancyocid[key] = fmt.Sprintf("%v", value)
					log.DefaultLogger.Error(q.tenancyocid[key])
				case "region":
					q.region[key] = fmt.Sprintf("%v", value)
				case "user":
					q.user[key] = fmt.Sprintf("%v", value)
				case "privkey":
					q.privkey[key] = fmt.Sprintf("%v", value)
				case "fingerprint":
					q.fingerprint[key] = fmt.Sprintf("%v", value)
				case "privkeypass":
					q.privkeypass[key] = EmptyKeyPass
				}
			}
		} else {
			TenancySettingsBlock++
			FieldIndex--
		}
	}
	return q, nil
}
