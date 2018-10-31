package metrics

import (
	"net/http"

	"github.com/oracle/oci-go-sdk/common"
)

type ListMetricsDetails struct {

	// The metric name to use when searching for metric definitions.
	// Example: `CpuUtilization`
	Name *string `mandatory:"false" json:"name"`

	// The source service or application to use when searching for metric definitions.
	// Example: `oci_computeagent`
	Namespace *string `mandatory:"false" json:"namespace"`

	// Qualifiers that you want to use when searching for metric definitions.
	// Available dimensions vary by metric namespace. Each dimension takes the form of a key-value pair.
	// Example: { "resourceId": "<var>&lt;instance_OCID&gt;</var>" }
	DimensionFilters map[string]string `mandatory:"false" json:"dimensionFilters"`

	// Group metrics by these fields in the response. For example, to list all metric namespaces available
	// in a compartment, groupBy the "namespace" field.
	// Example - group by namespace and resource:
	// `[ "namespace", "resourceId" ]`
	GroupBy []string `mandatory:"false" json:"groupBy"`

	// The field to use when sorting returned metric definitions. Only one sorting level is provided.
	// Example: `NAMESPACE`
	SortBy ListMetricsDetailsSortByEnum `mandatory:"false" json:"sortBy,omitempty"`

	// The sort order to use when sorting returned metric definitions. Ascending (ASC) or
	// descending (DESC).
	// Example: `ASC`
	SortOrder ListMetricsDetailsSortOrderEnum `mandatory:"false" json:"sortOrder,omitempty"`
}

func (m ListMetricsDetails) String() string {
	return common.PointerString(m)
}

// ListMetricsDetailsSortByEnum Enum with underlying type: string
type ListMetricsDetailsSortByEnum string

// Set of constants representing the allowable values for ListMetricsDetailsSortByEnum
const (
	ListMetricsDetailsSortByNamespace ListMetricsDetailsSortByEnum = "NAMESPACE"
	ListMetricsDetailsSortByName      ListMetricsDetailsSortByEnum = "NAME"
)

var mappingListMetricsDetailsSortBy = map[string]ListMetricsDetailsSortByEnum{
	"NAMESPACE": ListMetricsDetailsSortByNamespace,
	"NAME":      ListMetricsDetailsSortByName,
}

// GetListMetricsDetailsSortByEnumValues Enumerates the set of values for ListMetricsDetailsSortByEnum
func GetListMetricsDetailsSortByEnumValues() []ListMetricsDetailsSortByEnum {
	values := make([]ListMetricsDetailsSortByEnum, 0)
	for _, v := range mappingListMetricsDetailsSortBy {
		values = append(values, v)
	}
	return values
}

// ListMetricsDetailsSortOrderEnum Enum with underlying type: string
type ListMetricsDetailsSortOrderEnum string

// Set of constants representing the allowable values for ListMetricsDetailsSortOrderEnum
const (
	ListMetricsDetailsSortOrderAsc  ListMetricsDetailsSortOrderEnum = "ASC"
	ListMetricsDetailsSortOrderDesc ListMetricsDetailsSortOrderEnum = "DESC"
)

var mappingListMetricsDetailsSortOrder = map[string]ListMetricsDetailsSortOrderEnum{
	"ASC":  ListMetricsDetailsSortOrderAsc,
	"DESC": ListMetricsDetailsSortOrderDesc,
}

// GetListMetricsDetailsSortOrderEnumValues Enumerates the set of values for ListMetricsDetailsSortOrderEnum
func GetListMetricsDetailsSortOrderEnumValues() []ListMetricsDetailsSortOrderEnum {
	values := make([]ListMetricsDetailsSortOrderEnum, 0)
	for _, v := range mappingListMetricsDetailsSortOrder {
		values = append(values, v)
	}
	return values
}

type ListMetricsRequest struct {

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the
	// resources monitored by the metric that you are searching for. Use tenancyId to search in
	// the root compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The dimensions used to filter metrics.
	ListMetricsDetails `contributesTo:"body"`

	// Customer part of the request identifier token. If you need to contact Oracle about a particular
	// request, please provide the complete request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List" call.
	// For important details about how pagination works, see List Pagination (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated "List" call.
	// 1 is the minimum, 1000 is the maximum.
	// For important details about how pagination works, see List Pagination (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Default: 1000
	// Example: 500
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// When true, returns resources from all compartments and subcompartments. The parameter can
	// only be set to true when compartmentId is the tenancy OCID (the tenancy is the root compartment).
	// A true value requires the user to have tenancy-level permissions. If this requirement is not met,
	// then the call is rejected. When false, returns resources from only the compartment specified in
	// compartmentId. Default is false.
	CompartmentIdInSubtree *bool `mandatory:"false" contributesTo:"query" name:"compartmentIdInSubtree"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListMetricsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListMetricsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListMetricsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListMetricsResponse wrapper for the ListMetrics operation
type ListMetricsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []Metric instances
	Items []Metric `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// For important details about how pagination works, see List Pagination (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListMetricsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListMetricsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

type Metric struct {

	// The name of the metric.
	// Example: `CpuUtilization`
	Name *string `mandatory:"false" json:"name"`

	// The source service or application emitting the metric.
	// Example: `oci_computeagent`
	Namespace *string `mandatory:"false" json:"namespace"`

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing
	// the resources monitored by the metric.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// Qualifiers provided in a metric definition. Available dimensions vary by metric namespace.
	// Each dimension takes the form of a key-value pair.
	// Example: "resourceId": "<var>&lt;instance_OCID&gt;</var>"
	Dimensions map[string]string `mandatory:"false" json:"dimensions"`
}

func (m Metric) String() string {
	return common.PointerString(m)
}
