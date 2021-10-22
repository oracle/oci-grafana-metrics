package client

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/healthchecks"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCIHealthChecks struct {
	ctx               context.Context
	healthCheckClient healthchecks.HealthChecksClient
}

func (ohc *OCIHealthChecks) GetHealthChecksTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_healthchecks", "GetHealthChecksTagsPerRegion", "Fetching the health check resource tags from the oci")

	var fetchedResourceDetails []healthchecks.PingMonitorSummary
	var pageHeader string

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := healthchecks.ListPingMonitorsRequest{
		CompartmentId: common.String(compartmentOCID),
		SortBy:        healthchecks.ListPingMonitorsSortByDisplayname,
	}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := ohc.healthCheckClient.ListPingMonitors(ohc.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_healthchecks", "GetHealthChecksTagsPerRegion", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	for _, item := range fetchedResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})

		resourceLabels[*item.Id] = map[string]string{
			"resource_name":  *item.DisplayName,
			"hc_home_region": *item.HomeRegion,
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}
