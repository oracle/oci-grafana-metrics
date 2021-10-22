package client

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/loadbalancer"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCILoadBalancer struct {
	ctx    context.Context
	client loadbalancer.LoadBalancerClient
}

func (olb *OCILoadBalancer) GetLBaaSResourceTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_loadbalancer", "GetLBaaSResourceTagsPerRegion", "Fetching the load balancer resource tags from the oci")

	var fetchedResourceDetails []loadbalancer.LoadBalancer
	var pageHeader string

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := loadbalancer.ListLoadBalancersRequest{
		CompartmentId:  common.String(compartmentOCID),
		Detail:         common.String("full"),
		SortBy:         loadbalancer.ListLoadBalancersSortByDisplayname,
		LifecycleState: loadbalancer.LoadBalancerLifecycleStateActive,
	}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := olb.client.ListLoadBalancers(olb.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_loadbalancer", "GetLBaaSResourceTagsPerRegion", err)
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

		lbType := "public"
		if *item.IsPrivate {
			lbType = "private"
		}

		resourceLabels[*item.Id] = map[string]string{
			"resource_name":  *item.DisplayName,
			"lb_shape":       *item.ShapeName,
			"lb_access_type": lbType,
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}
