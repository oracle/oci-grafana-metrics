package client

import (
	"context"
	"strconv"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/healthchecks"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCIHealthChecks struct {
	ctx               context.Context
	healthCheckClient healthchecks.HealthChecksClient
}

func (ohc *OCIHealthChecks) GetHealthChecksTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_healthchecks", "GetHealthChecksTagsPerRegion", "Fetching the health check resource tags from the oci")

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	var pageHeader string
	var allCompartmentData sync.Map
	var wg sync.WaitGroup

	// fetching data per compartment
	for _, compartmentInAction := range compartments {
		wg.Add(1)

		go func(resource models.OCIResource) {
			defer wg.Done()

			var fetchedPmDetails []healthchecks.PingMonitorSummary

			req := healthchecks.ListPingMonitorsRequest{
				CompartmentId: common.String(resource.OCID),
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

				fetchedPmDetails = append(fetchedPmDetails, resp.Items...)
				if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
					pageHeader = *resp.OpcNextPage
				} else {
					break
				}
			}

			allCompartmentData.Store(resource.Name, fetchedPmDetails)
		}(compartmentInAction)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentData.Range(func(key, value interface{}) bool {
		compartmentName := key.(string)
		fetchedPmData := value.([]healthchecks.PingMonitorSummary)

		for _, item := range fetchedPmData {
			resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
				ResourceID:   *item.Id,
				ResourceName: *item.DisplayName,
				DefinedTags:  item.DefinedTags,
				FreeFormTags: item.FreeformTags,
			})

			resourceLabels[*item.Id] = map[string]string{
				"resource_name":  *item.DisplayName,
				"compartment":    compartmentName,
				"hc_home_region": *item.HomeRegion,
				"hc_protocol":    string(item.Protocol),
				"hc_interval":    strconv.Itoa(*item.IntervalInSeconds),
			}
		}

		return true
	})

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}
