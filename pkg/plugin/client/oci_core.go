package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/core"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCICore struct {
	ctx                  context.Context
	computeClient        core.ComputeClient
	virtualNetworkClient core.VirtualNetworkClient
}

func (oc *OCICore) GetComputeResourceTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_core", "GetComputeResourceTagsPerRegion", "Fetching the compute instanse tags from the oci")

	var fetchedResourceDetails []core.Instance
	var pageHeader string

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := core.ListInstancesRequest{
		CompartmentId:  common.String(compartmentOCID),
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

	for _, item := range fetchedResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})

		resourceLabels[*item.Id] = map[string]string{
			"resource_name":                *item.DisplayName,
			"compute_launch_mode":          string(item.LaunchMode),
			"compute_memory":               fmt.Sprintf("%gGB", *item.ShapeConfig.MemoryInGBs),
			"compute_ocpus":                fmt.Sprintf("%g", *item.ShapeConfig.Ocpus),
			"compute_networking_bandwidth": fmt.Sprintf("%gGbps", *item.ShapeConfig.NetworkingBandwidthInGbps),
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

func (oc *OCICore) GetVNicResourceTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_core", "GetVNicResourceTagsPerRegion", "Fetching the attched vnic resource tags from the oci")

	var fetchedResourceDetails []core.VnicAttachment
	var pageHeader string

	computeResourceLabelChan := make(chan map[string]string)
	vcnResourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := core.ListVnicAttachmentsRequest{
		CompartmentId: common.String(compartmentOCID),
	}

	// pulling compute instance details to attach the parent instance name
	go func(crl chan map[string]string) {
		var fetchedResourceDetails []core.Instance
		var pageHeader string

		cResourceLabels := map[string]string{}

		computeReq := core.ListInstancesRequest{
			CompartmentId: common.String(compartmentOCID),
			SortBy:        core.ListInstancesSortByDisplayname,
		}

		for {
			if len(pageHeader) != 0 {
				req.Page = common.String(pageHeader)
			}

			resp, err := oc.computeClient.ListInstances(oc.ctx, computeReq)
			if err != nil {
				backend.Logger.Error("client.oci_core", "GetVNicResourceTagsPerRegion", err)
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
			backend.Logger.Error("client.oci_core", "GetVNicResourceTagsPerRegion", err)
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
			"vnic_physical_nic_id": strconv.Itoa(*item.NicIndex),
			"vnic_parent_instance": computeResourceLabels[*item.InstanceId],
			"vnic_state":           string(item.LifecycleState),
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, vcnResourceLabels
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
