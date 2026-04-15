// Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package plugin

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"sync"

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

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

const MaxPagesToFetch = 20
const SingleTenancyKey = "DEFAULT/"
const NoTenancy = "NoTenancy"

var EmptyString string = ""
var EmptyKeyPass *string = &EmptyString

type TenancyAccess struct {
	monitoringClient monitoring.MonitoringClient
	identityClient   identity.IdentityClient
	config           common.ConfigurationProvider
	regionClients    map[string]*monitoring.MonitoringClient
	mu               sync.RWMutex
	customDomain     string // Sovereign cloud domain for endpoint fallback if SDK region discovery fails
}

// GetMonitoringClientForRegion returns a monitoring client configured for the specified region.
// Clients are lazily created and cached. Empty region or ALL_REGION returns the default client.
func (ta *TenancyAccess) GetMonitoringClientForRegion(region string) (*monitoring.MonitoringClient, error) {
	if region == "" || region == constants.ALL_REGION {
		return &ta.monitoringClient, nil
	}

	region = strings.ToLower(region)

	ta.mu.RLock()
	client, ok := ta.regionClients[region]
	ta.mu.RUnlock()

	if ok {
		return client, nil
	}

	ta.mu.Lock()
	defer ta.mu.Unlock()

	// Double-check after acquiring write lock
	if client, ok := ta.regionClients[region]; ok {
		return client, nil
	}

	newClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(ta.config)
	if err != nil {
		return nil, fmt.Errorf("error creating monitoring client for region %s: %w", region, err)
	}

	rp := clientRetryPolicy()
	newClient.Configuration.RetryPolicy = &rp
	newClient.SetRegion(region)

	// Safety net: if SDK region discovery failed at startup, SetRegion may resolve
	// sovereign cloud regions to the wrong domain (oraclecloud.com). Verify and
	// fall back to manual endpoint construction using the configured custom domain.
	if ta.customDomain != "" && !strings.Contains(newClient.Host, ta.customDomain) {
		fallbackHost := "https://telemetry." + region + "." + ta.customDomain
		backend.Logger.Warn("GetMonitoringClientForRegion", "FallbackEndpoint",
			"SDK resolved wrong domain for region "+region+", using fallback: "+fallbackHost)
		newClient.Host = fallbackHost
	}

	ta.regionClients[region] = &newClient
	return &newClient, nil
}

type OCIDatasource struct {
	tenancyAccess map[string]*TenancyAccess
	logger        log.Logger
	nameToOCID    map[string]string
	// timeCacheUpdated time.Time
	backend.CallResourceHandler
	// clients  *client.OCIClients
	settings *models.OCIDatasourceSettings
	cache    *ristretto.Cache
}

type OCIConfigFile struct {
	tenancyocid  map[string]string
	region       map[string]string
	user         map[string]string
	fingerprint  map[string]string
	privkey      map[string]string
	privkeypass  map[string]*string
	customregion map[string]string
	customdomain map[string]string
	logger       log.Logger
}

type OCISecuredSettings struct {
	Profile_0      string `json:"profile0,omitempty"`
	Tenancy_0      string `json:"tenancy0,omitempty"`
	Region_0       string `json:"region0,omitempty"`
	User_0         string `json:"user0,omitempty"`
	Privkey_0      string `json:"privkey0,omitempty"`
	Fingerprint_0  string `json:"fingerprint0,omitempty"`
	CustomRegion_0 string `json:"customregion0,omitempty"`
	CustomDomain_0 string `json:"customdomain0,omitempty"`

	Profile_1      string `json:"profile1,omitempty"`
	Tenancy_1      string `json:"tenancy1,omitempty"`
	Region_1       string `json:"region1,omitempty"`
	User_1         string `json:"user1,omitempty"`
	Fingerprint_1  string `json:"fingerprint1,omitempty"`
	Privkey_1      string `json:"privkey1,omitempty"`
	CustomRegion_1 string `json:"customregion1,omitempty"`
	CustomDomain_1 string `json:"customdomain1,omitempty"`

	Profile_2      string `json:"profile2,omitempty"`
	Tenancy_2      string `json:"tenancy2,omitempty"`
	Region_2       string `json:"region2,omitempty"`
	User_2         string `json:"user2,omitempty"`
	Fingerprint_2  string `json:"fingerprint2,omitempty"`
	Privkey_2      string `json:"privkey2,omitempty"`
	CustomRegion_2 string `json:"customregion2,omitempty"`
	CustomDomain_2 string `json:"customdomain2,omitempty"`

	Profile_3      string `json:"profile3,omitempty"`
	Tenancy_3      string `json:"tenancy3,omitempty"`
	Region_3       string `json:"region3,omitempty"`
	User_3         string `json:"user3,omitempty"`
	Fingerprint_3  string `json:"fingerprint3,omitempty"`
	Privkey_3      string `json:"privkey3,omitempty"`
	CustomRegion_3 string `json:"customregion3,omitempty"`
	CustomDomain_3 string `json:"customdomain3,omitempty"`

	Profile_4      string `json:"profile4,omitempty"`
	Tenancy_4      string `json:"tenancy4,omitempty"`
	Region_4       string `json:"region4,omitempty"`
	User_4         string `json:"user4,omitempty"`
	Fingerprint_4  string `json:"fingerprint4,omitempty"`
	Privkey_4      string `json:"privkey4,omitempty"`
	CustomRegion_4 string `json:"customregion4,omitempty"`
	CustomDomain_4 string `json:"customdomain4,omitempty"`

	Profile_5      string `json:"profile5,omitempty"`
	Tenancy_5      string `json:"tenancy5,omitempty"`
	Region_5       string `json:"region5,omitempty"`
	User_5         string `json:"user5,omitempty"`
	Fingerprint_5  string `json:"fingerprint5,omitempty"`
	Privkey_5      string `json:"privkey5,omitempty"`
	CustomRegion_5 string `json:"customregion5,omitempty"`
	CustomDomain_5 string `json:"customdomain5,omitempty"`

	Xtenancy_0 string `json:"xtenancy0,omitempty"`
}

// NewOCIConfigFile - constructor
func NewOCIConfigFile() *OCIConfigFile {
	return &OCIConfigFile{
		tenancyocid:  make(map[string]string),
		region:       make(map[string]string),
		user:         make(map[string]string),
		fingerprint:  make(map[string]string),
		privkey:      make(map[string]string),
		privkeypass:  make(map[string]*string),
		customregion: make(map[string]string),
		customdomain: make(map[string]string),
		logger:       log.DefaultLogger,
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

// NewOCIDatasource creates a new instance of OCIDatasource with the provided settings.
// It initializes the datasource settings, config provider, and cache, and registers HTTP routes.
//
// Parameters:
//   - settings: backend.DataSourceInstanceSettings containing the datasource instance settings.
//
// Returns:
//   - instancemgmt.Instance: The created OCIDatasource instance.
//   - error: An error if the datasource creation fails.
func NewOCIDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	backend.Logger.Debug("NewOCIDatasource", "instanceID", settings.ID)

	o := NewOCIDatasourceConstructor()
	dsSettings := &models.OCIDatasourceSettings{}

	if err := dsSettings.Load(settings); err != nil {
		backend.Logger.Error("failed to load datasource settings", "method", "NewOCIDatasource", "error", err)
		return nil, err
	}
	o.settings = dsSettings

	backend.Logger.Debug("datasource settings loaded", "environment", dsSettings.Environment, "tenancyMode", dsSettings.TenancyMode)

	if len(o.tenancyAccess) == 0 {
		err := o.getConfigProvider(dsSettings.Environment, dsSettings.TenancyMode, settings)
		if err != nil {
			return nil, err
		}
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
		Metrics:     false,
	})
	if err != nil {
		backend.Logger.Error("failed to create cache", "method", "NewOCIDatasource", "error", err)
		return nil, err
	}
	o.cache = cache

	mux := http.NewServeMux()
	o.registerRoutes(mux)
	o.CallResourceHandler = httpadapter.New(mux)

	return o, nil
}

/**
* @Description: QueryData handles the execution of multiple queries against the OCI Monitoring service.
* It iterates through each query in the request, executes it, and aggregates the results into a single response.
*
* @param ctx context.Context - The context for the request.
* @param req *backend.QueryDataRequest - The request containing multiple queries to be executed.
*
* @return *backend.QueryDataResponse - The response containing the results of all executed queries.
* @return error - An error if any of the queries fail or if there are issues processing the request.
*
* @Details:
*   - This function is the main entry point for handling data queries in the OCI Grafana plugin.
*   - It receives a `QueryDataRequest` which can contain multiple individual queries.
*   - For each query, it calls the `query` method to execute the query and get the result.
*   - The results are then stored in a `QueryDataResponse` using the query's `RefID` as the key.
*   - The function logs the start of the query process and returns the aggregated response.
*   - If any error occurs during the query execution, it will be returned.
 */
func (o *OCIDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	backend.Logger.Debug("QueryData", "datasource", req.PluginContext.DataSourceInstanceSettings.Name)

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
	backend.Logger.Debug("CheckHealth", "pluginID", req.PluginContext.PluginID)

	hRes := &backend.CheckHealthResult{}

	if err := o.TestConnectivity(ctx); err != nil {
		hRes.Status = backend.HealthStatusError
		hRes.Message = err.Error()
		backend.Logger.Error("health check failed", "method", "CheckHealth", "error", err)
		return hRes, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Success",
	}, nil
}

// OCILoadSettings loads and processes OCI configuration settings from the Grafana data source instance settings.
//
// This function handles both secured and non-secured settings, merging them to create a comprehensive
// configuration. It iterates through the settings, parsing and storing them in an OCIConfigFile struct.
// The function supports multiple tenancy configurations, identified by a numerical suffix (e.g., _0, _1).
//
// Parameters:
//   - req: backend.DataSourceInstanceSettings - The data source instance settings from Grafana.
//
// Returns:
//   - *OCIConfigFile: A pointer to an OCIConfigFile struct containing the parsed settings.
//   - error: An error if any issues occur during the loading or parsing of settings.
//
// The function performs the following steps:
//  1. Initializes an empty OCIConfigFile.
//  2. Unmarshals the JSON data from req.JSONData into both OCISecuredSettings and OCIDatasourceSettings structs.
//  3. Merges the non-secured settings into the secured settings.
//  4. Iterates the six profile blocks (_0.._5) in order, using each non-empty Profile name as the map key.
//  5. Stores the tenancy OCID, region, user, private key, fingerprint, custom region, and custom domain in the OCIConfigFile.
//  6. Stops at the first profile block whose Profile name is empty.
//  7. Returns the populated OCIConfigFile or an error if any step fails.
func OCILoadSettings(req backend.DataSourceInstanceSettings) (*OCIConfigFile, error) {
	q := NewOCIConfigFile()

	// Load secured and non-secured settings
	var dat OCISecuredSettings
	var nonsecdat models.OCIDatasourceSettings

	if err := json.Unmarshal(req.JSONData, &dat); err != nil {
		return nil, fmt.Errorf("can not read Secured settings: %s", err.Error())
	}

	if err := json.Unmarshal(req.JSONData, &nonsecdat); err != nil {
		return nil, fmt.Errorf("can not read settings: %s", err.Error())
	}

	// merge non secured settings into secured
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

	dat.Xtenancy_0 = nonsecdat.Xtenancy_0

	dat.CustomRegion_0 = nonsecdat.CustomRegion_0
	dat.CustomRegion_1 = nonsecdat.CustomRegion_1
	dat.CustomRegion_2 = nonsecdat.CustomRegion_2
	dat.CustomRegion_3 = nonsecdat.CustomRegion_3
	dat.CustomRegion_4 = nonsecdat.CustomRegion_4
	dat.CustomRegion_5 = nonsecdat.CustomRegion_5

	blocks := [6]struct {
		Profile, Tenancy, Region, User, Privkey, Fingerprint, CustomRegion, CustomDomain string
	}{
		{dat.Profile_0, dat.Tenancy_0, dat.Region_0, dat.User_0, dat.Privkey_0, dat.Fingerprint_0, dat.CustomRegion_0, dat.CustomDomain_0},
		{dat.Profile_1, dat.Tenancy_1, dat.Region_1, dat.User_1, dat.Privkey_1, dat.Fingerprint_1, dat.CustomRegion_1, dat.CustomDomain_1},
		{dat.Profile_2, dat.Tenancy_2, dat.Region_2, dat.User_2, dat.Privkey_2, dat.Fingerprint_2, dat.CustomRegion_2, dat.CustomDomain_2},
		{dat.Profile_3, dat.Tenancy_3, dat.Region_3, dat.User_3, dat.Privkey_3, dat.Fingerprint_3, dat.CustomRegion_3, dat.CustomDomain_3},
		{dat.Profile_4, dat.Tenancy_4, dat.Region_4, dat.User_4, dat.Privkey_4, dat.Fingerprint_4, dat.CustomRegion_4, dat.CustomDomain_4},
		{dat.Profile_5, dat.Tenancy_5, dat.Region_5, dat.User_5, dat.Privkey_5, dat.Fingerprint_5, dat.CustomRegion_5, dat.CustomDomain_5},
	}

	for _, b := range blocks {
		if b.Profile == "" {
			break
		}
		key := b.Profile
		q.tenancyocid[key] = b.Tenancy
		q.region[key] = b.Region
		q.user[key] = b.User
		q.privkey[key] = b.Privkey
		q.fingerprint[key] = b.Fingerprint
		q.customregion[key] = b.CustomRegion
		q.customdomain[key] = b.CustomDomain
	}

	return q, nil
}

// getConfigProvider configures the OCI Datasource based on the provided environment and tenancy mode.
// It supports two environments: "local" and "OCI Instance".
//
// Parameters:
// - environment: A string indicating the environment type ("local" or "OCI Instance").
// - tenancymode: A string indicating the tenancy mode ("singletenancy" or "multitenancy").
// - req: A backend.DataSourceInstanceSettings object containing the datasource instance settings.
//
// Returns:
// - error: An error object if any issues occur during configuration, otherwise nil.
//
// For "local" environment:
// - Configures using User Principals.
// - Loads settings from the provided datasource instance settings.
// - Validates the PEM key.
// - Creates OCI monitoring and identity clients with retry policies.
// - Overrides region and domain if a custom region is configured.
// - Stores the configured clients in the tenancyAccess map.
//
// For "OCI Instance" environment:
// - Configures using Instance Principal.
// - Optionally configures cross-tenancy instance principal if Xtenancy_0 is set.
// - Creates OCI monitoring and identity clients.
// - Stores the configured clients in the tenancyAccess map.
//
// Returns an error if the environment type is unknown or if any configuration steps fail.
func (o *OCIDatasource) getConfigProvider(environment string, tenancymode string, req backend.DataSourceInstanceSettings) error {

	switch environment {
	case "local":
		log.DefaultLogger.Debug("Configuring using User Principals")

		q, err := OCILoadSettings(req)
		if err != nil {
			return errors.New("Error Loading config settings")
		}
		for key := range q.tenancyocid {

			var configProvider common.ConfigurationProvider
			if tenancymode != "multitenancy" {
				if key != "DEFAULT" {
					backend.Logger.Debug("single tenancy mode, skipping additional profile", "profile", key)
					continue
				}
			}
			// test if PEM key is valid
			block, _ := pem.Decode([]byte(q.privkey[key]))
			if block == nil {
				return errors.New("Invalid Private Key in profile " + key)
			}
			// Override region in Configuration Provider in case a Custom region is configured
			if q.customregion[key] != "" {
				backend.Logger.Debug("using custom region", "method", "getConfigProvider", "customRegion", q.customregion[key])
				configProvider = common.NewRawConfigurationProvider(q.tenancyocid[key], q.user[key], q.customregion[key], q.fingerprint[key], q.privkey[key], q.privkeypass[key])

				// Register custom home region in SDK's global region map so that
				// SetRegion() can resolve sovereign cloud endpoints correctly
				if q.customdomain[key] != "" {
					regionSchema := map[string]string{
						"regionIdentifier":     q.customregion[key],
						"realmKey":             q.customregion[key],
						"realmDomainComponent": q.customdomain[key],
						"regionKey":            q.customregion[key],
					}
					common.AddRegionSchemaForPlc(regionSchema)
				}
			} else {
				configProvider = common.NewRawConfigurationProvider(q.tenancyocid[key], q.user[key], q.region[key], q.fingerprint[key], q.privkey[key], q.privkeypass[key])
			}

			// creating oci monitoring client
			mrp := clientRetryPolicy()
			monitoringClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
			if err != nil {
				backend.Logger.Error("failed to create monitoring client", "method", "getConfigProvider", "profile", key, "error", err)
				return errors.New("error with client")
			}
			monitoringClient.Configuration.RetryPolicy = &mrp

			// creating oci identity client
			irp := clientRetryPolicy()
			identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
			if err != nil {
				return errors.New("Error creating identity client")
			}
			identityClient.Configuration.RetryPolicy = &irp

			tenancyocid, err := configProvider.TenancyOCID()
			if err != nil {
				return errors.New("error with TenancyOCID")
			}
			ta := &TenancyAccess{
				monitoringClient: monitoringClient,
				identityClient:   identityClient,
				config:           configProvider,
				regionClients:    make(map[string]*monitoring.MonitoringClient),
				customDomain:     q.customdomain[key],
			}
			if tenancymode == "multitenancy" {
				o.tenancyAccess[key+"/"+tenancyocid] = ta
			} else {
				o.tenancyAccess[SingleTenancyKey] = ta
			}

			// Discover and register all subscribed sovereign cloud regions so that
			// SetRegion() resolves correct endpoints for cross-region queries.
			// Note: AddRegionSchemaForPlc is idempotent (skips already-registered regions),
			// so we register unconditionally to avoid StringToRegion side effects (IMDS, disk I/O).
			if q.customregion[key] != "" && q.customdomain[key] != "" {
				listReq := identity.ListRegionSubscriptionsRequest{TenancyId: common.String(tenancyocid)}
				response, regErr := identityClient.ListRegionSubscriptions(context.Background(), listReq)
				if regErr != nil {
					backend.Logger.Warn("getConfigProvider", "RegionDiscovery", "Failed to discover subscribed regions: "+regErr.Error())
				} else {
					for _, item := range response.Items {
						if item.RegionName == nil || *item.RegionName == "" {
							continue
						}
						regionName := *item.RegionName
						// Register unconditionally — AddRegionSchemaForPlc is a no-op
						// for regions already in the SDK's global map
						schema := map[string]string{
							"regionIdentifier":     regionName,
							"realmKey":             regionName,
							"realmDomainComponent": q.customdomain[key],
							"regionKey":            regionName,
						}
						common.AddRegionSchemaForPlc(schema)
						backend.Logger.Warn("getConfigProvider", "RegionDiscovery", "Registered sovereign region: "+regionName)
					}
				}
			}
		}
		return nil

	case "OCI Instance":
		log.DefaultLogger.Debug("Configuring using Instance Principal")
		var configProvider common.ConfigurationProvider
		configProvider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return errors.New("error with instance principals")
		}
		if o.settings.Xtenancy_0 != "" {
			log.DefaultLogger.Debug("Configuring using Cross Tenancy Instance Principal")
			tocid, _ := configProvider.TenancyOCID()
			log.DefaultLogger.Debug("Source Tenancy OCID: " + tocid)
			log.DefaultLogger.Debug("Target Tenancy OCID: " + o.settings.Xtenancy_0)
		}
		monitoringClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
		if err != nil {
			backend.Logger.Error("failed to create monitoring client", "method", "getConfigProvider", "profile", SingleTenancyKey, "error", err)
			return errors.New("error with client")
		}
		identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
		if err != nil {
			return errors.New("Error creating identity client")
		}
		o.tenancyAccess[SingleTenancyKey] = &TenancyAccess{
			monitoringClient: monitoringClient,
			identityClient:   identityClient,
			config:           configProvider,
			regionClients:    make(map[string]*monitoring.MonitoringClient),
		}
		return nil

	default:
		return errors.New("unknown environment type")
	}
}
