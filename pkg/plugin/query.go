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

	// checking if the query has valid Interval detail
	if qm.Interval == "" {
		backend.Logger.Warn("plugin.query", "query", "Interval is mandatory but it is not present in query")
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
		LegendFormat:    qm.LegendFormat,
		TagsValues:      qm.TagsValues,
		StartTime:       query.TimeRange.From.UTC(),
		EndTime:         query.TimeRange.To.UTC(),
	}

	// create data frame response
	frame := data.NewFrame("response").SetMeta(&data.FrameMeta{ExecutedQueryString: qm.QueryText})

	times, metricDataValues := ocidx.GetMetricDataPoints(ctx, metricsDataRequest, qm.TenancyOCID)

	// plotting the x axis with time as unit
	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	var name string
	for _, metricDataValue := range metricDataValues {
		name = metricDataValue.ResourceName
		ocidx.logger.Debug("UniqueDataID", "UniqueDataID", metricDataValue.UniqueDataID)

		dl := data.Labels{
			"tenancy":   metricDataValue.TenancyName,
			"unique_id": metricDataValue.UniqueDataID,
			"region":    metricDataValue.Region,
		}

		if qm.LegendFormat != "" {
			dl = data.Labels{}
			dimensions := ocidx.GetDimForLabel(ctx, qm.TenancyOCID, qm.CompartmentOCID, qm.Region, qm.Namespace, metricDataValue.MetricName)
			OriginalDimensionMap := make(map[string][]string)
			FoundDimensionMap := make(map[string][]string)
			var index int

			// convert dimension in a go map
			for _, dimension := range dimensions {
				key := dimension.Key
				ocidx.logger.Debug("KEY DIM", "key", key)
				for _, vall := range dimension.Values {
					OriginalDimensionMap[key] = dimension.Values
					ocidx.logger.Debug("ALL DIM", "dim", vall)
				}
			}

			//Search for resourceID and mark the position if found
			for _, value := range OriginalDimensionMap {
				for i, v := range value {
					if v == metricDataValue.UniqueDataID {
						index = i
						break
					}
				}
			}

			// Create a new map containing only the dimensions for the found resourceID
			for key, value := range OriginalDimensionMap {
				if len(value) > index {
					FoundDimensionMap[key] = []string{value[index]}
				}
			}

			name = ocidx.generateCustomMetricLabel(metricsDataRequest.LegendFormat, metricDataValue.MetricName, FoundDimensionMap)

		} else {
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
		}
		frame.Fields = append(frame.Fields,
			data.NewField(name, dl, metricDataValue.DataPoints),
		)
	}

	// add the frames to the response
	response.Frames = append(response.Frames, frame)

	return response
}
