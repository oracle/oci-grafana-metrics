// Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/dgraph-io/ristretto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/client"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
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

type OCIDatasource struct {
	tenancyAccess    map[string]*TenancyAccess
	logger           log.Logger
	nameToOCID       map[string]string
	timeCacheUpdated time.Time
	backend.CallResourceHandler
	clients  *client.OCIClients
	settings *models.OCIDatasourceSettings
	cache    *ristretto.Cache
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

func NewOCIDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	backend.Logger.Debug("plugin", "NewOCIDatasource", settings.ID)
	// var ts GrafanaCommonRequest

	o := NewOCIDatasourceConstructor()
	dsSettings := &models.OCIDatasourceSettings{}

	if err := dsSettings.Load(settings); err != nil {
		backend.Logger.Error("plugin", "NewOCIDatasource", "failed to load oci datasource settings: "+err.Error())
		return nil, err
	}
	o.settings = dsSettings

	backend.Logger.Error("plugin", "dsSettings.Environment", "dsSettings.Environment: "+dsSettings.Environment)
	backend.Logger.Error("plugin", "dsSettings.TenancyMode", "dsSettings.TenancyMode: "+dsSettings.TenancyMode)

	if len(o.tenancyAccess) == 0 {
		err := o.getConfigProvider(dsSettings.Environment, dsSettings.TenancyMode, settings)
		if err != nil {
			return nil, errors.New("broken environment")
		}
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
		Metrics:     false,
	})
	if err != nil {
		backend.Logger.Error("plugin", "NewOCIDatasource", "failed to create cache: "+err.Error())
		return nil, err
	}
	o.cache = cache

	// ociClients, err := client.New(dsSettings, cache)
	// if err != nil {
	// 	backend.Logger.Error("plugin", "NewOCIDatasource", "failed to load oci client: "+err.Error())
	// 	return nil, err
	// }
	// o.clients = ociClients

	mux := http.NewServeMux()
	o.registerRoutes(mux)
	o.CallResourceHandler = httpadapter.New(mux)

	// if err := json.Unmarshal(settings.JSONData, &ts); err != nil {
	// 	return nil, errors.New("can not read settings")
	// }

	return o, nil
}

func (o *OCIDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	backend.Logger.Debug("plugin", "QueryData", req.PluginContext.DataSourceInstanceSettings.Name)

	var ts GrafanaCommonRequest
	var takey string

	query := req.Queries[0]
	if err := json.Unmarshal(query.JSON, &ts); err != nil {
		return &backend.QueryDataResponse{}, err
	}
	tenancymode := o.settings.TenancyMode
	// queryType := ts.QueryType

	if tenancymode == "multitenancy" {
		takey = ts.Tenancy
	} else {
		takey = SingleTenancyKey
	}

	if len(o.tenancyAccess) == 0 {
		return &backend.QueryDataResponse{
			Responses: backend.Responses{
				query.RefID: backend.DataResponse{
					Error: fmt.Errorf("no such tenancy access key %q, make sure your datasources are migrated", takey),
				},
			},
		}, nil
	}
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := o.query(ctx, req.PluginContext, q)

		// saving the response in a hashmap based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// CheckHealth Handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (o *OCIDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	backend.Logger.Debug("plugin", "CheckHealth", req.PluginContext.PluginID)

	hRes := &backend.CheckHealthResult{}

	if err := o.TestConnectivity(ctx); err != nil {
		hRes.Status = backend.HealthStatusError
		hRes.Message = err.Error()
		backend.Logger.Error("plugin", "CheckHealth", err)

		return hRes, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Success",
	}, nil
}

// OCILoadSettings will read and validate Settings from the DataSourceConfig
func OCILoadSettings(req backend.DataSourceInstanceSettings) (*OCIConfigFile, error) {
	q := NewOCIConfigFile()

	TenancySettingsBlock := 0
	var dat OCISecuredSettings
	var nonsecdat models.OCIDatasourceSettings

	if err := json.Unmarshal(req.JSONData, &dat); err != nil {
		return nil, fmt.Errorf("can not read Secured settings: %s", err.Error())
	}

	if err := json.Unmarshal(req.JSONData, &nonsecdat); err != nil {
		return nil, fmt.Errorf("can not read settings: %s", err.Error())
	}

	decryptedJSONData := req.DecryptedSecureJSONData
	transcode(decryptedJSONData, &dat)

	dat.Region_0 = nonsecdat.Region_0
	dat.Region_1 = nonsecdat.Region_1
	dat.Region_2 = nonsecdat.Region_2
	dat.Region_3 = nonsecdat.Region_3
	dat.Region_4 = nonsecdat.Region_4
	dat.Region_5 = nonsecdat.Region_5

	dat.Profile_0 = nonsecdat.Profile_0
	dat.Profile_1 = nonsecdat.Profile_1
	dat.Profile_2 = nonsecdat.Profile_2
	dat.Profile_3 = nonsecdat.Profile_3
	dat.Profile_4 = nonsecdat.Profile_4
	dat.Profile_5 = nonsecdat.Profile_5

	log.DefaultLogger.Error(dat.Region_0)
	log.DefaultLogger.Error(nonsecdat.Region_0)

	log.DefaultLogger.Error(dat.Profile_0)
	log.DefaultLogger.Error(nonsecdat.Profile_0)

	v := reflect.ValueOf(dat)
	typeOfS := v.Type()
	var key string

	backend.Logger.Error("OCILoadSettings", "OCILoadSettings", "Siamo in OCILoadSettings")

	for FieldIndex := 0; FieldIndex < v.NumField(); FieldIndex++ {
		backend.Logger.Error("OCILoadSettings2", "OCILoadSettings2", typeOfS.Field(FieldIndex).Name)
		splits := strings.Split(typeOfS.Field(FieldIndex).Name, "_")
		SettingsBlockIndex, interr := strconv.Atoi(splits[1])
		if interr != nil {
			return nil, fmt.Errorf("can not read settings: %s", interr.Error())
		}
		backend.Logger.Error("OCILoadSettings3", "OCILoadSettings3", splits[0])
		backend.Logger.Error("OCILoadSettings4", "OCILoadSettings4", splits[1])

		if SettingsBlockIndex == TenancySettingsBlock {
			if splits[0] == "Profile" {
				backend.Logger.Error("asasas", "asasas", v.Field(FieldIndex).Interface())
				if v.Field(FieldIndex).Interface() != "" {
					key = fmt.Sprintf("%v", v.Field(FieldIndex).Interface())
					backend.Logger.Error("key", "key", key)

				} else {
					backend.Logger.Error("keyelse", "keyelse", v.Field(FieldIndex).Interface())
					return q, nil
				}
			} else {

				switch value := v.Field(FieldIndex).Interface(); strings.ToLower(splits[0]) {
				case "tenancy":
					q.tenancyocid[key] = fmt.Sprintf("%v", value)
					backend.Logger.Error("tenancy", "tenancy", value)
				case "region":
					q.region[key] = fmt.Sprintf("%v", value)
					backend.Logger.Error("region", "region", value)
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

func (o *OCIDatasource) getConfigProvider(environment string, tenancymode string, req backend.DataSourceInstanceSettings) error {

	// TEST statements
	var dat OCISecuredSettings
	decryptedJSONData := req.DecryptedSecureJSONData
	transcode(decryptedJSONData, &dat)
	log.DefaultLogger.Error(environment)
	log.DefaultLogger.Error(tenancymode)

	log.DefaultLogger.Error(dat.Tenancy_0)
	log.DefaultLogger.Error(dat.Tenancy_1)
	log.DefaultLogger.Error(dat.Tenancy_2)
	log.DefaultLogger.Error(dat.Tenancy_3)
	log.DefaultLogger.Error(dat.Tenancy_4)
	log.DefaultLogger.Error(dat.Tenancy_5)

	log.DefaultLogger.Error(dat.Region_0)
	log.DefaultLogger.Error(dat.Region_1)
	log.DefaultLogger.Error(dat.Region_2)
	log.DefaultLogger.Error(dat.Region_3)
	log.DefaultLogger.Error(dat.Region_4)
	log.DefaultLogger.Error(dat.Region_5)

	log.DefaultLogger.Error(dat.User_0)
	log.DefaultLogger.Error(dat.User_1)
	log.DefaultLogger.Error(dat.User_2)
	log.DefaultLogger.Error(dat.User_3)
	log.DefaultLogger.Error(dat.User_4)
	log.DefaultLogger.Error(dat.User_5)

	log.DefaultLogger.Error(dat.Profile_0)
	log.DefaultLogger.Error(dat.Profile_1)
	log.DefaultLogger.Error(dat.Profile_2)
	log.DefaultLogger.Error(dat.Profile_3)
	log.DefaultLogger.Error(dat.Profile_4)
	log.DefaultLogger.Error(dat.Profile_5)

	log.DefaultLogger.Error(dat.Fingerprint_0)
	log.DefaultLogger.Error(dat.Fingerprint_1)
	log.DefaultLogger.Error(dat.Fingerprint_2)
	log.DefaultLogger.Error(dat.Fingerprint_3)
	log.DefaultLogger.Error(dat.Fingerprint_4)
	log.DefaultLogger.Error(dat.Fingerprint_5)

	log.DefaultLogger.Error(dat.Privkey_0)
	log.DefaultLogger.Error(dat.Privkey_1)
	log.DefaultLogger.Error(dat.Privkey_2)
	log.DefaultLogger.Error(dat.Privkey_3)
	log.DefaultLogger.Error(dat.Privkey_4)
	log.DefaultLogger.Error(dat.Privkey_5)

	// end test statements

	switch environment {
	case "oci-user-principals":
		log.DefaultLogger.Error("User Principals siamo qui")
		q, err := OCILoadSettings(req)
		if err != nil {
			return errors.New("Error Loading config settings")
		}
		for key, _ := range q.tenancyocid {
			log.DefaultLogger.Error("Key: " + key)
			var configProvider common.ConfigurationProvider
			configProvider = common.NewRawConfigurationProvider(q.tenancyocid[key], q.user[key], q.region[key], q.fingerprint[key], q.privkey[key], q.privkeypass[key])
			metricsClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
			if err != nil {
				backend.Logger.Error("Error with config:" + key)
				return errors.New("error with client")
			}
			identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
			if err != nil {
				return errors.New("Error creating identity client")
			}
			tenancyocid, err := configProvider.TenancyOCID()
			if err != nil {
				return errors.New("error with TenancyOCID")
			}
			if tenancymode == "multitenancy" {
				o.tenancyAccess[key+"/"+tenancyocid] = &TenancyAccess{metricsClient, identityClient, configProvider}
			} else {
				o.tenancyAccess[SingleTenancyKey] = &TenancyAccess{metricsClient, identityClient, configProvider}
			}
		}
		return nil

	case "oci-instance":
		var configProvider common.ConfigurationProvider
		configProvider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return errors.New("error with instance principals")
		}
		metricsClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
		if err != nil {
			backend.Logger.Error("Error with config:" + SingleTenancyKey)
			return errors.New("error with client")
		}
		identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
		if err != nil {
			return errors.New("Error creating identity client")
		}
		o.tenancyAccess[SingleTenancyKey] = &TenancyAccess{metricsClient, identityClient, configProvider}
		return nil

	default:
		return errors.New("unknown environment type")
	}
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
			backend.Logger.Debug(key, "SKIPPED", fmt.Sprintf("listMetrics on Tenancy %s did not work, testing compartments", tenancyocid))
			comparts, Comperr := o.getCompartments(ctx, tenancyocid, regio, key)
			if Comperr != nil {
				return errors.Wrap(Comperr, fmt.Sprintf("error fetching Compartments"))
			}

			for _, v := range comparts {
				backend.Logger.Debug(key, "Testing", v)
				listMetrics := monitoring.ListMetricsRequest{
					CompartmentId: common.String(v),
				}

				res, err := o.tenancyAccess[key].metricsClient.ListMetrics(ctx, listMetrics)
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

func (o *OCIDatasource) getCompartments(ctx context.Context, rootCompartment string, region string, takey string) (map[string]string, error) {
	m := make(map[string]string)

	tenancyOcid := rootCompartment

	reg := common.StringToRegion(region)
	o.tenancyAccess[takey].metricsClient.SetRegion(string(reg))
	req := identity.GetTenancyRequest{TenancyId: common.String(tenancyOcid)}

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

/*
Function generates an array  containing OCI tenancy list in the following format:
<Label/TenancyOCID>
*/
func (o *OCIDatasource) GetTenancies2(ctx context.Context) []models.OCIResource {
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
	return subscribedRegions
}

func (o *OCIDatasource) GetTenancyAccessKey(tenancyOCID string) string {

	var takey string
	tenancymode := o.settings.TenancyMode

	if tenancymode == "multitenancy" {
		takey = tenancyOCID
	} else {
		takey = SingleTenancyKey
	}
	return takey
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

	// // fetching from cache, if present
	// cacheKey := strings.Join([]string{tenancyOCID, "cs"}, "-")
	// if cachedCompartments, found := oc.cache.Get(cacheKey); found {
	// 	backend.Logger.Warn("client", "GetCompartments", "getting the data from cache")
	// 	return cachedCompartments.([]models.OCIResource)
	// }

	takey := o.GetTenancyAccessKey(tenancyOCID)
	var tenancyocid string
	var tenancyErr error

	tenancymode := o.settings.TenancyMode

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

	// regio, regErr := o.tenancyAccess[takey].config.Region()
	// if regErr != nil {
	// 	return nil
	// }

	compartments := map[string]string{}
	// calling the api if not present in cache
	compartmentList := []models.OCIResource{}
	var fetchedCompartments []identity.Compartment
	var pageHeader string

	for {
		// reg := common.StringToRegion(regio)
		// o.tenancyAccess[takey].metricsClient.SetRegion(string(reg))
		req := identity.ListCompartmentsRequest{
			CompartmentId:          common.String(tenancyocid),
			CompartmentIdInSubtree: common.Bool(true),
			LifecycleState:         identity.CompartmentLifecycleStateActive,
			Limit:                  common.Int(1000),
		}

		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		res, err := o.tenancyAccess[takey].identityClient.ListCompartments(ctx, req)
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

	// // saving in the cache
	// oc.cache.SetWithTTL(cacheKey, compartmentList, 1, 15*time.Minute)
	// oc.cache.Wait()

	return compartmentList
}
