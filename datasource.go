package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gitlab-odx.oracledx.com/cloudnative/oci-grafana-plugin/metrics"

	"golang.org/x/net/context"

	"github.com/davecgh/go-spew/spew"
	"github.com/grafana/grafana_plugin_model/go/datasource"
	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/pkg/errors"
)

//how often to refresh our compartmentID cache
var cacheRefreshTime = time.Minute

//OCIDatasource - pulls in data from telemtry/various oci apis
type OCIDatasource struct {
	plugin.NetRPCUnsupportedPlugin
	metricsClient    metrics.TelemetryClient
	identityClient   identity.IdentityClient
	config           common.ConfigurationProvider
	logger           hclog.Logger
	nameToOCID       map[string]string
	timeCacheUpdated time.Time
}

//NewOCIDatasource - constructor
func NewOCIDatasource(pluginLogger hclog.Logger) (*OCIDatasource, error) {
	m := make(map[string]string)

	return &OCIDatasource{
		logger:     pluginLogger,
		nameToOCID: m,
	}, nil
}

// GrafanaOCIRequest - Query Request comning in from the front end
type GrafanaOCIRequest struct {
	GrafanaCommonRequest
	Query      string
	Resolution string
	Namespace  string
}

//GrafanaSearchRequest incoming request body for search requests
type GrafanaSearchRequest struct {
	GrafanaCommonRequest
	Metric    string `json:"metric,omitempty"`
	Namespace string
}

type GrafanaCompartmentRequest struct {
	GrafanaCommonRequest
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
func (o *OCIDatasource) Query(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	o.logger.Debug("Query", "datasource", tsdbReq.Datasource.Name, "TimeRange", tsdbReq.TimeRange)
	var ts GrafanaCommonRequest
	json.Unmarshal([]byte(tsdbReq.Queries[0].ModelJson), &ts)

	queryType := tsdbReq.Queries[0].RefId
	if o.config == nil {
		configProvider, err := getConfigProvider(ts.Environment)
		if err != nil {
			return nil, errors.Wrap(err, "broken environment")
		}
		metricsClient, err := metrics.NewTelemetryClientWithConfigurationProvider(configProvider)
		if err != nil {
			return nil, errors.New(fmt.Sprint("error with client", spew.Sdump(configProvider), err.Error()))
		}
		identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
		if err != nil {
			log.Printf("error with client")
			panic(err)
		}
		o.identityClient = identityClient
		o.metricsClient = metricsClient
		o.config = configProvider
	}

	switch queryType {
	case "compartments":
		return o.compartmentsResponse(ctx, tsdbReq)
	case "dimensions":
		return o.dimensionResponse(ctx, tsdbReq)
	case "namespaces":
		return o.namespaceResponse(ctx, tsdbReq)
	case "search":
		return o.searchResponse(ctx, tsdbReq)
	case "test":
		return o.testResponse(ctx, tsdbReq)
	default:
		return o.queryResponse(ctx, tsdbReq)
	}
}

func (o *OCIDatasource) testResponse(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	var ts GrafanaCommonRequest
	json.Unmarshal([]byte(tsdbReq.Queries[0].ModelJson), &ts)

	listMetrics := metrics.ListMetricsRequest{
		CompartmentId: common.String(ts.TenancyOCID),
	}
	reg := common.StringToRegion(ts.Region)
	o.metricsClient.SetRegion(string(reg))
	res, err := o.metricsClient.ListMetrics(ctx, listMetrics)
	status := res.RawResponse.StatusCode
	if status >= 200 && status < 300 {
		return &datasource.DatasourceResponse{}, nil
	}
	return nil, errors.Wrap(err, fmt.Sprintf("list metrircs failed %s %d", spew.Sdump(res), status))
}

func (o *OCIDatasource) dimensionResponse(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	table := datasource.Table{
		Columns: []*datasource.TableColumn{
			&datasource.TableColumn{Name: "text"},
		},
		Rows: make([]*datasource.TableRow, 0),
	}

	for _, query := range tsdbReq.Queries {
		var ts GrafanaSearchRequest
		json.Unmarshal([]byte(query.ModelJson), &ts)
		reqDetails := metrics.ListMetricsDetails{}
		reqDetails.Namespace = common.String(ts.Namespace)
		reqDetails.Name = common.String(ts.Metric)
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint("list metrircs failed", spew.Sdump(reqDetails)))
		}
		rows := make([]*datasource.TableRow, 0)
		for _, item := range items {
			for dimension, value := range item.Dimensions {
				rows = append(rows, &datasource.TableRow{
					Values: []*datasource.RowValue{
						&datasource.RowValue{
							Kind:        datasource.RowValue_TYPE_STRING,
							StringValue: fmt.Sprintf("%s=%s", dimension, value),
						},
					},
				})
			}
		}
		table.Rows = rows
	}
	return &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			&datasource.QueryResult{
				RefId:  "dimensions",
				Tables: []*datasource.Table{&table},
			},
		},
	}, nil
}

func (o *OCIDatasource) namespaceResponse(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	table := datasource.Table{
		Columns: []*datasource.TableColumn{
			&datasource.TableColumn{Name: "text"},
		},
		Rows: make([]*datasource.TableRow, 0),
	}
	for _, query := range tsdbReq.Queries {
		var ts GrafanaSearchRequest
		json.Unmarshal([]byte(query.ModelJson), &ts)

		reqDetails := metrics.ListMetricsDetails{}
		reqDetails.GroupBy = []string{"namespace"}
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint("list metrircs failed", spew.Sdump(reqDetails)))
		}

		rows := make([]*datasource.TableRow, 0)
		for _, item := range items {
			rows = append(rows, &datasource.TableRow{
				Values: []*datasource.RowValue{
					&datasource.RowValue{
						Kind:        datasource.RowValue_TYPE_STRING,
						StringValue: *(item.Namespace),
					},
				},
			})
		}
		table.Rows = rows
	}
	return &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			&datasource.QueryResult{
				RefId:  "namespaces",
				Tables: []*datasource.Table{&table},
			},
		},
	}, nil
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

func (o *OCIDatasource) searchResponse(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	table := datasource.Table{
		Columns: []*datasource.TableColumn{
			&datasource.TableColumn{Name: "text"},
		},
		Rows: make([]*datasource.TableRow, 0),
	}

	for _, query := range tsdbReq.Queries {
		var ts GrafanaSearchRequest
		json.Unmarshal([]byte(query.ModelJson), &ts)
		reqDetails := metrics.ListMetricsDetails{
			Namespace: common.String(ts.Namespace),
		}
		items, err := o.searchHelper(ctx, ts.Region, ts.Compartment, reqDetails)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprint("list metrircs failed", spew.Sdump(reqDetails)))
		}

		rows := make([]*datasource.TableRow, 0)
		metricCache := make(map[string]bool)
		for _, item := range items {
			if _, ok := metricCache[*(item.Name)]; !ok {
				rows = append(rows, &datasource.TableRow{
					Values: []*datasource.RowValue{
						&datasource.RowValue{
							Kind:        datasource.RowValue_TYPE_STRING,
							StringValue: *(item.Name),
						},
					},
				})
				metricCache[*(item.Name)] = true
			}
		}
		table.Rows = rows
	}
	return &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			&datasource.QueryResult{
				RefId:  "search",
				Tables: []*datasource.Table{&table},
			},
		},
	}, nil

}

func (o *OCIDatasource) searchHelper(ctx context.Context, region, compartment string, metricDetails metrics.ListMetricsDetails) ([]metrics.Metric, error) {
	var items []metrics.Metric
	var page *string
	for {
		reg := common.StringToRegion(region)
		o.metricsClient.SetRegion(string(reg))
		res, err := o.metricsClient.ListMetrics(ctx, metrics.ListMetricsRequest{
			CompartmentId:      common.String(compartment),
			ListMetricsDetails: metricDetails,
			Page:               page,
		})
		if err != nil {
			return nil, errors.Wrap(err, "list metrircs failed")
		}
		items = append(items, res.Items...)
		if res.OpcNextPage == nil {
			break
		}
		page = res.OpcNextPage
	}
	return items, nil
}

func (o *OCIDatasource) compartmentsResponse(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	table := datasource.Table{
		Columns: []*datasource.TableColumn{
			&datasource.TableColumn{Name: "text"},
			&datasource.TableColumn{Name: "text"},
		},
	}
	now := time.Now()
	var ts GrafanaSearchRequest
	json.Unmarshal([]byte(tsdbReq.Queries[0].ModelJson), &ts)
	if o.timeCacheUpdated.IsZero() || now.Sub(o.timeCacheUpdated) > cacheRefreshTime {
		o.logger.Debug("refreshing cache")
		m, err := o.getCompartments(ctx, ts.TenancyOCID)
		if err != nil {
			o.logger.Error("Unable to refresh cache")
			return nil, err
		}
		o.nameToOCID = m
	}

	rows := make([]*datasource.TableRow, 0, len(o.nameToOCID))
	for name, id := range o.nameToOCID {
		val := &datasource.RowValue{
			Kind:        datasource.RowValue_TYPE_STRING,
			StringValue: name,
		}
		id := &datasource.RowValue{
			Kind:        datasource.RowValue_TYPE_STRING,
			StringValue: id,
		}

		rows = append(rows, &datasource.TableRow{
			Values: []*datasource.RowValue{
				val,
				id,
			},
		})
	}
	table.Rows = rows
	return &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			&datasource.QueryResult{
				RefId:  "compartment",
				Tables: []*datasource.Table{&table},
			},
		},
	}, nil
}

func (o *OCIDatasource) getCompartments(ctx context.Context, rootCompartment string) (map[string]string, error) {
	m := make(map[string]string)
	m["root compartment"] = rootCompartment
	var page *string
	for {
		res, err := o.identityClient.ListCompartments(ctx,
			identity.ListCompartmentsRequest{
				CompartmentId: &rootCompartment,
				Page:          page,
			})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("this is what we were trying to get %s", rootCompartment))
		}
		for _, compartment := range res.Items {
			m[*(compartment.Name)] = *(compartment.Id)
		}
		if res.OpcNextPage == nil {
			break
		}
		page = res.OpcNextPage
	}
	return m, nil
}

type responseAndQuery struct {
	ociRes metrics.SummarizeMetricsDataResponse
	query  *datasource.Query
	err    error
}

func (o *OCIDatasource) queryResponse(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	results := make([]responseAndQuery, 0, len(tsdbReq.Queries))
	for _, query := range tsdbReq.Queries {
		var ts GrafanaOCIRequest
		json.Unmarshal([]byte(query.ModelJson), &ts)

		start := time.Unix(tsdbReq.TimeRange.FromEpochMs/1000, (tsdbReq.TimeRange.FromEpochMs%1000)*1000000).UTC()
		end := time.Unix(tsdbReq.TimeRange.ToEpochMs/1000, (tsdbReq.TimeRange.ToEpochMs%1000)*1000000).UTC()

		start = start.Truncate(time.Millisecond)
		end = end.Truncate(time.Millisecond)

		req := metrics.SummarizeMetricsDataDetails{
			Query:      common.String(ts.Query),
			Namespace:  common.String(ts.Namespace),
			StartTime:  &common.SDKTime{start},
			EndTime:    &common.SDKTime{end},
			Resolution: common.String(ts.Resolution),
		}
		reg := common.StringToRegion(ts.Region)
		o.metricsClient.SetRegion(string(reg))

		request := metrics.SummarizeMetricsDataRequest{
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
	queryRes := make([]*datasource.QueryResult, 0, len(results))
	for _, q := range results {
		res := &datasource.QueryResult{
			RefId: q.query.RefId,
		}
		if q.err != nil {
			res.Error = q.err.Error()
			queryRes = append(queryRes, res)
			continue
		}
		//Items -> timeserries
		series := make([]*datasource.TimeSeries, 0, len(q.ociRes.Items))
		for _, item := range q.ociRes.Items {
			t := &datasource.TimeSeries{
				Name: *(item.Name),
			}
			for k, v := range item.Dimensions {
				if k == "resourceId" {
					t.Name = fmt.Sprintf("%s, {%s}", t.Name, v)
				}
			}
			p := make([]*datasource.Point, 0, len(item.AggregatedDatapoints))
			for _, metric := range item.AggregatedDatapoints {
				point := &datasource.Point{
					Timestamp: int64(metric.Timestamp.UnixNano() / 1000000),
					Value:     *(metric.Value),
				}
				p = append(p, point)
			}
			t.Points = p
			series = append(series, t)
		}
		res.Series = series
		queryRes = append(queryRes, res)
	}

	response := &datasource.DatasourceResponse{
		Results: queryRes,
	}

	return response, nil
}
