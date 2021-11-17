package client

import (
	"context"
	"strconv"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v51/apmcontrolplane"
	"github.com/oracle/oci-go-sdk/v51/apmsynthetics"
	"github.com/oracle/oci-go-sdk/v51/common"
	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCIApm struct {
	ctx             context.Context
	domainClient    apmcontrolplane.ApmDomainClient
	syntheticClient apmsynthetics.ApmSyntheticClient
}

func (oa *OCIApm) getApmDomainTags(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_apm", "getApmDomainTags", "Fetching the apm domains for compartment: "+compartmentOCID)

	var fetchedResourceDetails []apmcontrolplane.ApmDomainSummary
	var pageHeader string

	apmDomainLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := apmcontrolplane.ListApmDomainsRequest{
		CompartmentId:  common.String(compartmentOCID),
		SortBy:         apmcontrolplane.ListApmDomainsSortByDisplayname,
		LifecycleState: apmcontrolplane.ListApmDomainsLifecycleStateActive,
		Limit:          common.Int(50),
	}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := oa.domainClient.ListApmDomains(oa.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_apm", "getApmDomainTags", err)
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

		apmDomainLabels[*item.Id] = map[string]string{
			"apm_domain_name": *item.DisplayName,
			"resource_name":   *item.DisplayName,
		}

		// apmDomainLabels[*item.Id] = map[string]string{
		// 	"apm_domain_name": *item.DisplayName,
		// 	"apm_domain_free": "false",
		// }

		// if item.IsFreeTier != nil && *item.IsFreeTier {
		// 	apmDomainLabels[*item.Id]["apm_domain_free"] = "true"
		// }
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, apmDomainLabels
}

func (oa *OCIApm) getApmMonitorLabelsPerDomain(apmDomainOCID string) map[string]map[string]string {
	backend.Logger.Debug("client.oci_apm", "getApmMonitorLabelsPerDomain", "Fetching the apm monitors for apm domain : "+apmDomainOCID)

	var fetchedResourceDetails []apmsynthetics.MonitorSummary
	var pageHeader string

	apmMonitorLabels := map[string]map[string]string{}

	req := apmsynthetics.ListMonitorsRequest{
		ApmDomainId: common.String(apmDomainOCID),
		Status:      apmsynthetics.ListMonitorsStatusEnabled,
		SortBy:      apmsynthetics.ListMonitorsSortByDisplayname,
		Limit:       common.Int(50),
	}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := oa.syntheticClient.ListMonitors(oa.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_apm", "getApmMonitorLabelsPerDomain", err)
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
		apmMonitorLabels[*item.Id] = map[string]string{
			"apm_monitor_name":             *item.DisplayName,
			"apm_monitor_type":             string(item.MonitorType),
			"apm_monitor_target":           *item.Target,
			"apm_monitor_repeat_interval":  strconv.Itoa(*item.RepeatIntervalInSeconds) + "s",
			"apm_monitor_timeout_interval": strconv.Itoa(*item.TimeoutInSeconds) + "s",
			//"apm_monitor_vantage_points_count": strconv.Itoa(*item.VantagePointCount),
		}

		if item.ScriptName != nil {
			apmMonitorLabels[*item.Id]["apm_monitor_script_name"] = *item.ScriptName
		}

		noOfVantagePoints := *item.VantagePointCount
		vantagePoints := ""
		for _, vp := range item.VantagePoints {
			vantagePoints += *vp.Name
			noOfVantagePoints -= 1

			if noOfVantagePoints != 0 {
				vantagePoints += ","
			}
		}

		apmMonitorLabels[*item.Id]["apm_monitor_vantage_points"] = vantagePoints
	}

	return apmMonitorLabels
}

func (oa *OCIApm) getApmTagsPerCompartment(compartment models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_apm", "getApmTagsPerCompartment", "Fetching the apm tags for compartment : "+compartment.Name)

	var wg sync.WaitGroup
	var allDomainMonitorData sync.Map
	resourceLabels := map[string]map[string]string{}

	// getting the list of apm domains and associated labels
	resourceTags, resourceIDsPerTag, apmDomainLabels := oa.getApmDomainTags(compartment.OCID)

	for apmDomainOCID, apmDomainMetaData := range apmDomainLabels {
		wg.Add(1)

		go func(ocid string, metaData map[string]string) {
			defer wg.Done()

			apmMonitorLabels := oa.getApmMonitorLabelsPerDomain(ocid)

			for monitorID, monitorMetaData := range apmMonitorLabels {
				// adding the apm domain meta data
				for k, v := range metaData {
					monitorMetaData[k] = v
				}

				// adding compartment name
				monitorMetaData["compartment"] = compartment.Name

				// re-assigning the key-value pairs
				apmMonitorLabels[monitorID] = monitorMetaData
			}

			allDomainMonitorData.Store(ocid, apmMonitorLabels)
		}(apmDomainOCID, apmDomainMetaData)
	}
	wg.Wait()

	// collecting the data from all apm domain
	allDomainMonitorData.Range(func(key, value interface{}) bool {
		apmDomainOCID := key.(string)
		apmMonitorLabelsPerDomain := value.(map[string]map[string]string)

		for apmMonitorOCID, apmMonitorLabels := range apmMonitorLabelsPerDomain {
			resourceLabels[apmDomainOCID+apmMonitorOCID] = apmMonitorLabels
		}

		return true
	})

	return resourceTags, resourceIDsPerTag, resourceLabels
}

func (oa *OCIApm) GetApmTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_apm", "GetApmTagsPerRegion", "Fetching the apm resource tags from the oci")

	// when queried for a single compartment
	if len(compartments) == 1 {
		return oa.getApmTagsPerCompartment(compartments[0])
	}

	return nil, nil, nil

	// var allCompartmentData sync.Map
	// var wg sync.WaitGroup

	// // fetching data per compartment
	// for _, compartmentInAction := range compartments {
	// 	wg.Add(1)

	// 	go func(compartment models.OCIResource) {
	// 		defer wg.Done()

	// 		resourceTags, resourceIDsPerTag, resourceLabels := oa.getApmTagsPerCompartment(compartment)

	// 	}(compartmentInAction)
	// }
	// wg.Wait()
}
