package client

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v48/common"
	"github.com/oracle/oci-go-sdk/v48/core"
	"github.com/oracle/oci-go-sdk/v48/identity"
	"github.com/oracle/oci-go-sdk/v48/loadbalancer"
	"github.com/oracle/oci-go-sdk/v48/monitoring"
)

// OCIClient stores all the clients related to OCI
type OCIClient struct {
	tenancyOCID          string
	region               string
	clientConfigProvider common.ConfigurationProvider
	identityClient       identity.IdentityClient
	monitoringClient     monitoring.MonitoringClient
	computeClient        core.ComputeClient
}

func (oc *OCIClient) GetComputeClient() (core.ComputeClient, error) {
	crp := clientRetryPolicy()
	// creating oci core compute client
	computeClient, err := core.NewComputeClientWithConfigurationProvider(oc.clientConfigProvider)
	if err != nil {
		backend.Logger.Error("client.oci_client", "GetComputeClient", "could not create oci core compute client: %v", err)
		return core.ComputeClient{}, err
	}

	computeClient.Configuration.RetryPolicy = &crp

	return computeClient, nil
}

func (oc *OCIClient) GetVCNClient() (core.VirtualNetworkClient, error) {
	crp := clientRetryPolicy()
	// creating oci core vcn client
	vcnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(oc.clientConfigProvider)
	if err != nil {
		backend.Logger.Error("client.oci_client", "GetVCNClient", "could not create oci core vcn client: %v", err)
		return core.VirtualNetworkClient{}, err
	}

	vcnClient.Configuration.RetryPolicy = &crp

	return vcnClient, nil
}

func (oc *OCIClient) GetLBaaSClient() (loadbalancer.LoadBalancerClient, error) {
	crp := clientRetryPolicy()
	// creating oci lbaas client
	lbaasClient, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(oc.clientConfigProvider)
	if err != nil {
		backend.Logger.Error("client.oci_client", "GetLBaaSClient", "could not create oci lbaas client: %v", err)
		return loadbalancer.LoadBalancerClient{}, err
	}

	lbaasClient.Configuration.RetryPolicy = &crp

	return lbaasClient, nil
}
