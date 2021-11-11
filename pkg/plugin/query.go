package plugin

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	jsoniter "github.com/json-iterator/go"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

func (ocidx *OCIDatasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	backend.Logger.Debug("plugin.query", "query", "query initiated for "+query.RefID)

	// Creating the Data response for query
	response := backend.DataResponse{}

	// Unmarshal the json into oci queryModel
	qm := &models.QueryModel{}
	response.Error = jsoniter.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// checking if the query has valid tenancy detail
	if qm.TenancyOCID == "" {
		backend.Logger.Warn("plugin.query", "query", "tenancy ocid is mandatory but it is not present in query")
		return response
	}

	metricsDataRequest := models.MetricsDataRequest{
		TenancyOCID:     qm.TenancyOCID,
		CompartmentOCID: qm.CompartmentOCID,
		CompartmentName: qm.CompartmentName,
		Region:          qm.Region,
		Namespace:       qm.Namespace,
		QueryText:       qm.QueryText,
		Interval:        qm.Interval[1 : len(qm.Interval)-1],
		ResourceGroup:   qm.ResourceGroup,
		DimensionValues: qm.DimensionValues,
		TagsValues:      qm.TagsValues,
		StartTime:       query.TimeRange.From.UTC(),
		EndTime:         query.TimeRange.To.UTC(),
	}

	// create data frame response
	frame := data.NewFrame("response").SetMeta(&data.FrameMeta{ExecutedQueryString: qm.QueryText})

	times, metricDataValues := ocidx.clients.GetMetricDataPoints(ctx, metricsDataRequest)

	// plotting the x axis with time as unit
	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))

	for _, metricDataValue := range metricDataValues {
		name := metricDataValue.ResourceName
		dl := data.Labels{
			"tenancy":   metricDataValue.TenancyName,
			"unique_id": metricDataValue.UniqueDataID,
			"region":    metricDataValue.Region,
		}

		for k, v := range metricDataValue.Labels {
			dl[k] = v
			if k == "resource_name" && len(name) == 0 {
				name = v
			}
			// if k != "resource_name" {
			// 	dl[k] = v
			// } else {
			// 	if len(name) == 0 {
			// 		name = v
			// 	}
			// }
		}

		frame.Fields = append(frame.Fields,
			data.NewField(name, dl, metricDataValue.DataPoints),
		)
	}

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}
