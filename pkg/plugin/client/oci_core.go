package client

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCICore struct {
	ctx                  context.Context
	computeClient        core.ComputeClient
	virtualNetworkClient core.VirtualNetworkClient
}

func (oc *OCICore) GetComputeResourceTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_core", "GetComputeResourceTagsPerRegion", "Fetching the compute instanse tags from the oci")

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

			var fetchedResourceDetails []core.Instance

			req := core.ListInstancesRequest{
				CompartmentId:  common.String(resource.OCID),
				SortBy:         core.ListInstancesSortByDisplayname,
				LifecycleState: core.InstanceLifecycleStateRunning,
			}

			for {
				if len(pageHeader) != 0 {
					req.Page = common.String(pageHeader)
				}

				resp, err := oc.computeClient.ListInstances(oc.ctx, req)
				if err != nil {
					backend.Logger.Error("client.oci_core", "GetComputeResourceTagsPerRegion", err)
					break
				}

				fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
				if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
					pageHeader = *resp.OpcNextPage
				} else {
					break
				}
			}

			allCompartmentData.Store(resource.Name, fetchedResourceDetails)
		}(compartmentInAction)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentData.Range(func(key, value interface{}) bool {
		compartmentName := key.(string)
		fetchedResourceData := value.([]core.Instance)

		for _, item := range fetchedResourceData {
			resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
				ResourceID:   *item.Id,
				ResourceName: *item.DisplayName,
				DefinedTags:  item.DefinedTags,
				FreeFormTags: item.FreeformTags,
			})

			resourceLabels[*item.Id] = map[string]string{
				"resource_name":                *item.DisplayName,
				"compartment":                  compartmentName,
				"compute_launch_mode":          string(item.LaunchMode),
				"compute_memory":               fmt.Sprintf("%gGB", *item.ShapeConfig.MemoryInGBs),
				"compute_ocpus":                fmt.Sprintf("%g", *item.ShapeConfig.Ocpus),
				"compute_networking_bandwidth": fmt.Sprintf("%gGbps", *item.ShapeConfig.NetworkingBandwidthInGbps),
			}
		}

		return true
	})

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

func (oc *OCICore) getVNicResourceTagsPerCompartment(compartment models.OCIResource) (map[string]map[string]struct{}, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_core", "getVNicResourceTagsPerCompartment", "Fetching the attched vnic resource tags from the oci per compartment")

	var fetchedResourceDetails []core.VnicAttachment
	var pageHeader string

	computeResourceLabelChan := make(chan map[string]string)
	vcnResourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := core.ListVnicAttachmentsRequest{
		CompartmentId: common.String(compartment.OCID),
	}

	// pulling compute instance details to attach the parent instance name
	go func(crl chan map[string]string) {
		var fetchedResourceDetails []core.Instance
		var pageHeader string

		cResourceLabels := map[string]string{}

		computeReq := core.ListInstancesRequest{
			CompartmentId: common.String(compartment.OCID),
			SortBy:        core.ListInstancesSortByDisplayname,
		}

		for {
			if len(pageHeader) != 0 {
				req.Page = common.String(pageHeader)
			}

			resp, err := oc.computeClient.ListInstances(oc.ctx, computeReq)
			if err != nil {
				backend.Logger.Error("client.oci_core", "getVNicResourceTagsPerCompartment:ListInstances", err)
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
			cResourceLabels[*item.Id] = *item.DisplayName
		}

		crl <- cResourceLabels
	}(computeResourceLabelChan)

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := oc.computeClient.ListVnicAttachments(oc.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_core", "getVNicResourceTagsPerCompartment:ListVnicAttachments", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	computeResourceLabels := <-computeResourceLabelChan

	for _, item := range fetchedResourceDetails {
		displayName := ""
		if item.DisplayName != nil {
			displayName = *item.DisplayName
		}
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.VnicId,
			ResourceName: displayName,
		})

		vcnResourceLabels[*item.VnicId] = map[string]string{
			"resource_name":        displayName,
			"compartment":          compartment.Name,
			"vnic_physical_nic_id": strconv.Itoa(*item.NicIndex),
			"vnic_parent_instance": computeResourceLabels[*item.InstanceId],
			"vnic_state":           string(item.LifecycleState),
		}
	}

	resourceTags, resourceIDsPerTag := collectResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, vcnResourceLabels
}

func (oc *OCICore) GetVNicResourceTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_core", "GetVNicResourceTagsPerRegion", "Fetching the attched vnic resource tags from the oci")

	// when queried for a single compartment
	if len(compartments) == 1 {
		resourceTags, resourceIDsPerTag, apmLabels := oc.getVNicResourceTagsPerCompartment(compartments[0])
		return convertToArray(resourceTags), resourceIDsPerTag, apmLabels
	}

	// holds key: value1, value2, for UI
	allResourceTags := map[string]map[string]struct{}{}
	// holds key.value: map of resourceIDs, for caching
	allResourceIDsPerTag := map[string]map[string]struct{}{}
	allResourceLabels := map[string]map[string]string{}

	var allCompartmentData sync.Map
	var wg sync.WaitGroup

	// fetching data per compartment
	for _, compartmentInAction := range compartments {
		wg.Add(1)

		go func(compartment models.OCIResource) {
			defer wg.Done()

			resourceTags, resourceIDsPerTag, resourceLabels := oc.getVNicResourceTagsPerCompartment(compartment)

			allCompartmentData.Store(compartment.OCID, map[string]interface{}{
				"resourceTags":      resourceTags,
				"resourceIDsPerTag": resourceIDsPerTag,
				"resourceLabels":    resourceLabels,
			})

		}(compartmentInAction)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentData.Range(func(key, value interface{}) bool {
		// compartmentOCID := key.(string)
		apmAllCompartmentData := value.(map[string]interface{})

		newResourceTags := apmAllCompartmentData["resourceTags"].(map[string]map[string]struct{})
		newResourceIDsPerTag := apmAllCompartmentData["resourceIDsPerTag"].(map[string]map[string]struct{})
		newResourceLabels := apmAllCompartmentData["resourceLabels"].(map[string]map[string]string)

		if len(allResourceTags) == 0 {
			allResourceTags = newResourceTags
			allResourceIDsPerTag = newResourceIDsPerTag
			allResourceLabels = newResourceLabels

			return true
		}

		// checking each new key and values, for resource tags
		for newTagKey, newTagValues := range newResourceTags {
			// when the key is already present in the collected
			if existingTagValues, ok := allResourceTags[newTagKey]; ok {
				// checking each new value in the collected ones
				for v := range newTagValues {
					// add it when not found
					if _, found := existingTagValues[v]; !found {
						existingTagValues[v] = struct{}{}
						allResourceTags[newTagKey] = existingTagValues
					}
				}
			} else {
				// for new key
				allResourceTags[newTagKey] = newTagValues
			}
		}

		// checking each new key and values, for resource ids
		for newTagKey, newTagValues := range newResourceIDsPerTag {
			// when the key is already present in the collected
			if existingTagValues, ok := allResourceIDsPerTag[newTagKey]; ok {
				// checking each new value in the collected ones
				for v := range newTagValues {
					// add it when not found
					if _, found := existingTagValues[v]; !found {
						existingTagValues[v] = struct{}{}
						allResourceIDsPerTag[newTagKey] = existingTagValues
					}
				}
			} else {
				// for new key
				allResourceIDsPerTag[newTagKey] = newTagValues
			}
		}

		// checking each new key and values, for resource labels
		for newResourceID, newResourceLabelValues := range newResourceLabels {
			// when the key is already present in the collected
			if _, ok := allResourceLabels[newResourceID]; !ok {
				allResourceLabels[newResourceID] = newResourceLabelValues
			}
		}

		return true
	})

	return convertToArray(allResourceTags), allResourceIDsPerTag, allResourceLabels
}

func (oc *OCICore) GetVCNResourceTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_core", "GetVCNResourceTagsPerRegion", "Fetching the vcn resource tags from the oci")

	var fetchedResourceDetails []core.Vcn
	var pageHeader string

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := core.ListVcnsRequest{
		CompartmentId:  common.String(compartmentOCID),
		SortBy:         core.ListVcnsSortByDisplayname,
		LifecycleState: core.VcnLifecycleStateAvailable,
	}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := oc.virtualNetworkClient.ListVcns(oc.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_core", "GetVCNResourceTagsPerRegion", err)
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
			"resource_name": *item.DisplayName,
			"vcn_domain":    *item.VcnDomainName,
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}
