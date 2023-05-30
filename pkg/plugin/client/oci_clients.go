package client

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v65/apmcontrolplane"
	"github.com/oracle/oci-go-sdk/v65/apmsynthetics"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/database"
	"github.com/oracle/oci-go-sdk/v65/healthchecks"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
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
		backend.Logger.Error("client.oci_client", "GetComputeClient", "could not create oci core compute client: "+err.Error())
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
		backend.Logger.Error("client.oci_client", "GetVCNClient", "could not create oci core vcn client: "+err.Error())
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
		backend.Logger.Error("client.oci_client", "GetLBaaSClient", "could not create oci lbaas client: "+err.Error())
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
		backend.Logger.Error("client.oci_client", "GetHealthChecksClient", "could not create oci health checks client: "+err.Error())
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
		backend.Logger.Error("client.oci_client", "GetDatabaseClient", "could not create oci database client: "+err.Error())
		return database.DatabaseClient{}, err
	}

	dbClient.Configuration.RetryPolicy = &crp

	return dbClient, nil
}

func (oc *OCIClient) GetApmClients() (apmcontrolplane.ApmDomainClient, apmsynthetics.ApmSyntheticClient, error) {
	crp := clientRetryPolicy()

	// creating oci apm domain client
	apmDomainClient, admErr := apmcontrolplane.NewApmDomainClientWithConfigurationProvider(oc.clientConfigProvider)
	if admErr != nil {
		backend.Logger.Error("client.oci_client", "GetApmClients", "could not create oci apm domain client: "+admErr.Error())
		return apmcontrolplane.ApmDomainClient{}, apmsynthetics.ApmSyntheticClient{}, admErr
	}
	apmDomainClient.Configuration.RetryPolicy = &crp

	// creating oci apm synthetic client
	apmSyntheticClient, asmErr := apmsynthetics.NewApmSyntheticClientWithConfigurationProvider(oc.clientConfigProvider)
	if asmErr != nil {
		backend.Logger.Error("client.oci_client", "GetApmClients", "could not create oci apm synthetic client: "+asmErr.Error())
		return apmDomainClient, apmsynthetics.ApmSyntheticClient{}, asmErr
	}
	apmSyntheticClient.Configuration.RetryPolicy = &crp

	return apmDomainClient, apmSyntheticClient, nil
}
