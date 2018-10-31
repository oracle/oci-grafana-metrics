package metrics

import (
	"net/http"

	"github.com/oracle/oci-go-sdk/common"
)

// SummarizeMetricsDataDetails The request details for retrieving aggregated data.
// Use the query and optional properties to filter the returned results.
// Maximum data points returned per call: 10,080
type SummarizeMetricsDataDetails struct {

	// The Telemetry Query Language (TQL) expression to use when searching for metric datapoints to
	// aggregate. The query must specify a metric, statistic, and interval. Supported values for
	// interval: `1m`, `5m`, `1h`. You can optionally specify dimensions and grouping functions.
	// Supported grouping functions: `grouping()`, `groupBy()`.
	// For available dimensions, review the metric definition.
	// Example: `CpuUtilization[1m].sum()`
	Query *string `mandatory:"true" json:"query"`

	// The source service or application to use when searching for metric data points to aggregate. If not
	// specified, then all available sources are used.
	// Example: `oci_computeagent`
	Namespace *string `mandatory:"false" json:"namespace"`

	// The beginning of the time range to use when searching for metric data points.
	// Format is defined by RFC3339. The response includes metric data points for the startTime.
	// Default value: the timestamp 3 hours before the call was sent.
	// Example: `2018-02-01T01:02:29.600Z`
	StartTime *common.SDKTime `mandatory:"false" json:"startTime"`

	// The end of the time range to use when searching for metric data points.
	// Format is defined by RFC3339. The response excludes metric data points for the endTime.
	// Default value: the timestamp representing when the call was sent.
	// Example: `2018-02-01T02:02:29.600Z`
	EndTime *common.SDKTime `mandatory:"false" json:"endTime"`

	// The time between calculated aggregation windows. Use with the query interval to vary the
	// frequency at which aggregated data points are returned. For example, use a query interval of
	// 5 minutes with a resolution of 1 minute to retrieve five-minute aggregations at a one-minute
	// frequency. The resolution must be equal or less than the interval in the query. The default
	// resolution is 1m (one minute). Supported values: `1m`, `5m`, `1h`.
	// Example: `5m`
	Resolution *string `mandatory:"false" json:"resolution"`
}

func (m SummarizeMetricsDataDetails) String() string {
	return common.PointerString(m)
}

// SummarizeMetricsDataRequest wrapper for the SummarizeMetricsData operation
type SummarizeMetricsDataRequest struct {

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the
	// resources monitored by the metric that you are searching for. Use tenancyId to search in
	// the root compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The dimensions used to filter for metrics.
	SummarizeMetricsDataDetails `contributesTo:"body"`

	// Customer part of the request identifier token. If you need to contact Oracle about a particular
	// request, please provide the complete request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request SummarizeMetricsDataRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request SummarizeMetricsDataRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request SummarizeMetricsDataRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// SummarizeMetricsDataResponse wrapper for the SummarizeMetricsData operation
type SummarizeMetricsDataResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []MetricData instance
	Items []MetricData `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response SummarizeMetricsDataResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response SummarizeMetricsDataResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

type MetricData struct {

	// The reference provided in a metric definition to indicate the source service or
	// application that emitted the metric.
	// Example: `oci_computeagent`
	Namespace *string `mandatory:"true" json:"namespace"`

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the
	// resources from which the aggregated data was returned.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The name of the metric.
	// Example: `CpuUtilization`
	Name *string `mandatory:"true" json:"name"`

	// Qualifiers provided in the definition of the returned metric.
	// Available dimensions vary by metric namespace. Each dimension takes the form of a key-value pair.
	// Example: `resourceId`
	Dimensions map[string]string `mandatory:"true" json:"dimensions"`

	// The list of timestamp-value pairs returned for the specified request. Metric values are rolled up to the start time specified in the request.
	AggregatedDatapoints []AggregatedDatapoint `mandatory:"true" json:"aggregatedDatapoints"`

	// The references provided in a metric definition to indicate extra information about the metric.
	// Example: `unit`
	Metadata map[string]string `mandatory:"false" json:"metadata"`

	// The time between calculated aggregation windows. Use with the query interval to vary the
	// frequency at which aggregated data points are returned. For example, use a query interval of
	// 5 minutes with a resolution of 1 minute to retrieve five-minute aggregations at a one-minute
	// frequency. The resolution must be equal or less than the interval in the query. The default
	// resolution is 1m (one minute). Supported values: `1m`, `5m`, `1h`.
	// Example: `5m`
	Resolution *string `mandatory:"false" json:"resolution"`
}

func (m MetricData) String() string {
	return common.PointerString(m)
}

type AggregatedDatapoint struct {

	// The date and time associated with the value of this data point. Format defined by RFC3339.
	// Example: `2018-02-01T01:02:29.600Z`
	Timestamp *common.SDKTime `mandatory:"true" json:"timestamp"`

	// Numeric value of the metric.
	// Example: `10.4`
	Value *float64 `mandatory:"true" json:"value"`
}

func (m AggregatedDatapoint) String() string {
	return common.PointerString(m)
}
