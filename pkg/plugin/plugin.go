package plugin

import (
	"context"
	"net/http"

	"github.com/dgraph-io/ristretto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/client"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCIDatasource struct {
	backend.CallResourceHandler
	clients  *client.OCIClients
	settings *models.OCIDatasourceSettings
	cache    *ristretto.Cache
}

func NewOCIDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	backend.Logger.Debug("plugin", "NewOCIDatasource", settings.ID)
	ociDx := &OCIDatasource{}
	dsSettings := &models.OCIDatasourceSettings{}

	if err := dsSettings.Load(settings); err != nil {
		backend.Logger.Error("plugin", "NewOCIDatasource", "failed to load oci datasource settings: "+err.Error())
		return nil, err
	}
	ociDx.settings = dsSettings

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
	ociDx.cache = cache

	ociClients, err := client.New(dsSettings, cache)
	if err != nil {
		backend.Logger.Error("plugin", "NewOCIDatasource", "failed to load oci client: "+err.Error())
		return nil, err
	}
	ociDx.clients = ociClients

	mux := http.NewServeMux()
	ociDx.registerRoutes(mux)
	ociDx.CallResourceHandler = httpadapter.New(mux)

	return ociDx, nil
}

// Dispose Called before creatinga a new instance to allow plugin authors
// to cleanup.
func (ocidx *OCIDatasource) Dispose() {
	backend.Logger.Debug("plugin", "NewOCIDatasource", "Clearing up")

	ocidx.clients.Destroy()
	ocidx.clients = nil
	ocidx.cache.Clear()
	ocidx.cache.Close()
}

// QueryData Primary method called by grafana-server to handle multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (ocidx *OCIDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	backend.Logger.Debug("plugin", "QueryData", req.PluginContext.DataSourceInstanceSettings.Name)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := ocidx.query(ctx, req.PluginContext, q)

		// saving the response in a hashmap based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// CheckHealth Handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (ocidx *OCIDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	backend.Logger.Debug("plugin", "CheckHealth", req.PluginContext.PluginID)

	hRes := &backend.CheckHealthResult{}

	if err := ocidx.clients.TestConnectivity(ctx); err != nil {
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
