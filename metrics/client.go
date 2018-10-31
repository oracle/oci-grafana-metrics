package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/oracle/oci-go-sdk/common"
)

//TelemetryClient a client for Telemetry
type TelemetryClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewTelemetryClientWithConfigurationProvider Creates a new default Telemetry client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewTelemetryClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client TelemetryClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = TelemetryClient{BaseClient: baseClient}
	client.BasePath = "20180401"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *TelemetryClient) SetRegion(region string) {
	client.Host = fmt.Sprintf(common.DefaultHostURLTemplate, "telemetry", region)
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *TelemetryClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
	if ok, err := common.IsConfigurationProviderValid(configProvider); !ok {
		return err
	}

	// Error has been checked already
	region, _ := configProvider.Region()
	client.SetRegion(region)
	client.config = &configProvider
	return nil
}

// ConfigurationProvider the ConfigurationProvider used in this client, or null if none set
func (client *TelemetryClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// SummarizeMetricsData Returns aggregated data that match the criteria specified in the request. Compartment OCID required.
// For more information on monitoring metrics, see Telemetry Overview (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/Telemetry/Concepts/telemetryoverview.htm).
func (client TelemetryClient) SummarizeMetricsData(ctx context.Context, request SummarizeMetricsDataRequest) (response SummarizeMetricsDataResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.summarizeMetricsData, policy)
	if err != nil {
		if ociResponse != nil {
			response = SummarizeMetricsDataResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(SummarizeMetricsDataResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into SummarizeMetricsDataResponse")
	}
	return
}

// summarizeMetricsData implements the OCIOperation interface (enables retrying operations)
func (client TelemetryClient) summarizeMetricsData(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/metrics/actions/summarizeMetricsData")
	if err != nil {
		log.Printf("error first line")
		return nil, err
	}

	var response SummarizeMetricsDataResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		log.Printf("error with call")
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListMetrics Returns metric definitions that match the criteria specified in the request. Compartment OCID required.
// For information about metrics, see Metrics Overview (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/Monitoring/Concepts/monitoringoverview.htm#MetricsOverview).
func (client TelemetryClient) ListMetrics(ctx context.Context, request ListMetricsRequest) (response ListMetricsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listMetrics, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListMetricsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListMetricsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListMetricsResponse")
	}
	return
}

// listMetrics implements the OCIOperation interface (enables retrying operations)
func (client TelemetryClient) listMetrics(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/metrics/actions/listMetrics")
	if err != nil {
		return nil, err
	}

	var response ListMetricsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}
