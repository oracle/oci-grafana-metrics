package client

import (
	"context"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v51/common"
	"github.com/oracle/oci-go-sdk/v51/loadbalancer"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCILoadBalancer struct {
	ctx    context.Context
	client loadbalancer.LoadBalancerClient
}

func (olb *OCILoadBalancer) GetLBaaSResourceTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_loadbalancer", "GetLBaaSResourceTagsPerRegion", "Fetching the load balancer resource tags from the oci")

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

			var fetchedLbDetails []loadbalancer.LoadBalancer

			req := loadbalancer.ListLoadBalancersRequest{
				CompartmentId:  common.String(resource.OCID),
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

			allCompartmentData.Store(resource.Name, fetchedLbDetails)
		}(compartmentInAction)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentData.Range(func(key, value interface{}) bool {
		compartmentName := key.(string)
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
				"compartment":    compartmentName,
				"lb_shape":       *item.ShapeName,
				"lb_access_type": lbType,
			}
		}

		return true
	})

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}
