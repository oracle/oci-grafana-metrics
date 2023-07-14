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

	ocidx.logger.Debug("000 Generated metric label", "LegendFormat", qm.LegendFormat)

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
		ocidx.logger.Debug("2 Dimensiona Label metric label", "name", name)
		for key, value := range metricDataValue.Labels {
			ocidx.logger.Debug("sciapo key", "sciapo key", key)
			ocidx.logger.Debug("sciapo value", "sciapo value", value)

		}

		if qm.LegendFormat != "" {
			dl = data.Labels{}
			dimensions := ocidx.GetDimForLabel(ctx, qm.TenancyOCID, qm.CompartmentOCID, qm.Region, qm.Namespace, metricDataValue.MetricName)
			ocidx.logger.Debug("aa Generated metric label", "legendFormat", qm.LegendFormat)
			// dims := map[string]string{}
			mymap := make(map[string][]string)
			newmap := make(map[string][]string)
			searchValue := metricDataValue.UniqueDataID
			var index int

			for _, dimension := range dimensions {
				key := dimension.Key
				ocidx.logger.Debug("KEY DIM", "key", key)
				for _, vall := range dimension.Values {
					mymap[key] = dimension.Values
					ocidx.logger.Debug("ALL DIM", "dim", vall)

				}
			}
			for _, value := range mymap {
				for i, v := range value {
					if v == searchValue {
						index = i
						break
					}
				}
			}

			for key, value := range mymap {
				if len(value) > index {
					newmap[key] = []string{value[index]}
				}
			}

			// for kuga, vuga := range dims {
			// 	ocidx.logger.Debug("KUGA DIM", "kuga", kuga)
			// 	ocidx.logger.Debug("VUGA DIM", "vuga", vuga)
			// }
			name = ocidx.generateCustomMetricLabel(metricsDataRequest.LegendFormat, metricDataValue.MetricName, newmap)
			// name = ocidx.OgenerateCustomMetricLabel(metricsDataRequest.LegendFormat, metricDataValue.MetricName, dims)

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
