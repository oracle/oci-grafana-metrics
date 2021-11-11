package client

import (
	"context"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/loadbalancer"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCILoadBalancer struct {
	ctx    context.Context
	client loadbalancer.LoadBalancerClient
}

func (olb *OCILoadBalancer) GetLBaaSResourceTagsPerRegion(compartmentOCIDs []string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_loadbalancer", "GetLBaaSResourceTagsPerRegion", "Fetching the load balancer resource tags from the oci")

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	var pageHeader string
	var allCompartmentData sync.Map
	var wg sync.WaitGroup

	// fetching data per compartment
	for _, compartmentOCID := range compartmentOCIDs {
		wg.Add(1)

		go func(ocid string) {
			defer wg.Done()

			var fetchedLbDetails []loadbalancer.LoadBalancer

			req := loadbalancer.ListLoadBalancersRequest{
				CompartmentId:  common.String(ocid),
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

				fetchedLbDetails = append(fetchedLbDetails, resp.Items...)
				if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
					pageHeader = *resp.OpcNextPage
				} else {
					break
				}
			}

			allCompartmentData.Store(ocid, fetchedLbDetails)
		}(compartmentOCID)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentData.Range(func(key, value interface{}) bool {
		fetchedLbData := value.([]loadbalancer.LoadBalancer)

		for _, item := range fetchedLbData {
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

		return true
	})

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}
