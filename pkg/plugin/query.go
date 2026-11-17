// Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package plugin

import (
	"context"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	jsoniter "github.com/json-iterator/go"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

// query handles the execution of a data query for the OCIDatasource plugin.
// It performs the following steps:
// 1. Logs the initiation of the query.
// 2. Creates a DataResponse object to hold the query results.
// 3. Unmarshals the JSON query into a QueryModel object.
// 4. Validates the presence of mandatory fields (TenancyOCID and Interval) in the query.
// 5. Constructs a MetricsDataRequest object with the necessary details for fetching metrics data.
// 6. Creates a data frame to store the response data.
// 7. Fetches metric data points if the region, compartment, and namespace are valid.
// 8. Handles errors during data fetching and returns the response if any errors occur.
// 9. Plots the x-axis with time as the unit and processes metric data values to populate the data frame.
// 10. Adds the data frame to the response and returns the response.
//
// Parameters:
// - ctx: The context for the query execution.
// - pCtx: The plugin context.
// - query: The data query to be executed.
//
// Returns:
// - backend.DataResponse: The response containing the query results or an error if the query fails.
func (ocidx *OCIDatasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	backend.Logger.Debug("query initiated", "method", "query", "refID", query.RefID)

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
		RawQuery:        qm.RawQuery,
		TagsValues:      qm.TagsValues,
		StartTime:       query.TimeRange.From.UTC(),
		EndTime:         query.TimeRange.To.UTC(),
	}

	// create data frame response
	frame := data.NewFrame("response").SetMeta(&data.FrameMeta{ExecutedQueryString: qm.QueryText})
	var err error
	var metricDataValues []models.OCIMetricDataPoints
	var times []time.Time

	if (qm.Region != "" && qm.Region != "select region") &&
		(qm.CompartmentOCID != "" && qm.CompartmentOCID != "select compartment") &&
		(qm.Namespace != "" && qm.Namespace != "select namespace") {
		times, metricDataValues, err = ocidx.GetMetricDataPoints(ctx, metricsDataRequest, qm.TenancyOCID)
	}
	if err != nil {
		response.Error = err
		return response
	}

	// plotting the x axis with time as unit
	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	var name string
	for _, metricDataValue := range metricDataValues {
		name = metricDataValue.ResourceName

		dl := data.Labels{
			"tenancy":   metricDataValue.TenancyName,
			"unique_id": metricDataValue.UniqueDataID,
			"region":    metricDataValue.Region,
		}
		if qm.LegendFormat != "" {
			// Use the dimensions OCI returned for this exact series. Re-fetching them via
			// ListMetrics and correlating by slice index was unreliable (issue #370).
			label, ok := ocidx.generateCustomMetricLabel(metricsDataRequest.LegendFormat, metricDataValue.MetricName, metricDataValue.Dimensions, dl)
			if ok {
				name = label
			} else if name == "" {
				// A placeholder could not be resolved for this series (for example
				// resourceDisplayName on a namespace that has no such dimension):
				// keep the default series name instead of a raw "{{...}}".
				name = metricDataValue.UniqueDataID
			}
			dl = data.Labels{}

		} else {
			for k, v := range metricDataValue.Labels {
				dl[k] = v
				if k == "resource_name" && len(name) == 0 {
					name = v
				}
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
