package client

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/core"
	"github.com/oracle/oci-go-sdk/v49/database"
	"github.com/oracle/oci-go-sdk/v49/healthchecks"
	"github.com/oracle/oci-go-sdk/v49/identity"
	"github.com/oracle/oci-go-sdk/v49/loadbalancer"
	"github.com/oracle/oci-go-sdk/v49/monitoring"
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

func (oc *OCIClient) GetHealthChecksClient() (healthchecks.HealthChecksClient, error) {
	crp := clientRetryPolicy()
	// creating oci health checks client
	hcClient, err := healthchecks.NewHealthChecksClientWithConfigurationProvider(oc.clientConfigProvider)
	if err != nil {
		backend.Logger.Error("client.oci_client", "GetHealthChecksClient", "could not create oci health checks client: %v", err)
		return healthchecks.HealthChecksClient{}, err
	}

	hcClient.Configuration.RetryPolicy = &crp

	return hcClient, nil
}

func (oc *OCIClient) GetDatabaseClient() (database.DatabaseClient, error) {
	crp := clientRetryPolicy()
	// creating oci database client
	dbClient, err := database.NewDatabaseClientWithConfigurationProvider(oc.clientConfigProvider)
	if err != nil {
		backend.Logger.Error("client.oci_client", "GetDatabaseClient", "could not create oci database client: %v", err)
		return database.DatabaseClient{}, err
	}

	dbClient.Configuration.RetryPolicy = &crp

	return dbClient, nil
}
