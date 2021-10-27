package client

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/database"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCIDatabase struct {
	ctx    context.Context
	client database.DatabaseClient
}

// getDatabaseHomes to fetch db home details
func (od *OCIDatabase) getDatabaseHomes(compartmentOCID string) []map[string]string {
	backend.Logger.Debug("client.oci_database", "getDatabaseHomes", "Fetching the database homes from the oci for compartment>"+compartmentOCID)

	var fetchedResourceDetails []database.DbHomeSummary
	var pageHeader string

	resourceInfo := []map[string]string{}

	req := database.ListDbHomesRequest{
		CompartmentId: common.String(compartmentOCID),
	}

	// backend.Logger.Debug("client.oci_database", "getDatabaseHomes", req)

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := od.client.ListDbHomes(od.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_database", "getDatabaseHomes", err)
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
		resourceInfo = append(resourceInfo, map[string]string{
			"db_home_id":   *item.Id,
			"db_home_name": *item.DisplayName,
			"db_system_id": *item.DbSystemId,
			"db_version":   *item.DbVersion,
		})
	}

	return resourceInfo
}

// GetDatabaseTagsPerRegion To fetch tags from an Oracle Database on a bare metal or virtual machine DB system.
func (od *OCIDatabase) GetDatabaseTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "GetDatabaseTagsPerRegion", "Fetching the database resource tags from the oci for compartment>"+compartmentOCID)

	var fetchedResourceDetails []database.DatabaseSummary
	var pageHeader string

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	// fetching the db homes
	dbHomes := od.getDatabaseHomes(compartmentOCID)

	for _, dbHome := range dbHomes {
		req := database.ListDatabasesRequest{
			CompartmentId: common.String(compartmentOCID),
			DbHomeId:      common.String(dbHome["db_home_id"]),
		}

		for {
			if len(pageHeader) != 0 {
				req.Page = common.String(pageHeader)
			}

			resp, err := od.client.ListDatabases(od.ctx, req)
			if err != nil {
				backend.Logger.Error("client.oci_database", "GetDatabaseTagsPerRegion", err)
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
				ResourceName: *item.DbName,
				DefinedTags:  item.DefinedTags,
				FreeFormTags: item.FreeformTags,
			})

			resourceLabels[*item.Id] = map[string]string{
				"resource_name":  *item.DbName,
				"db_name":        *item.DbName,
				"db_unique_name": *item.DbUniqueName,
				"pdb_name":       *item.PdbName,
				"db_home_name":   dbHome["db_home_name"],
				"db_version":     dbHome["db_version"],
			}
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

// GetAutonomousDatabaseTagsPerRegion To fetch tags from an Oracle Autonomous Database.
func (od *OCIDatabase) GetAutonomousDatabaseTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "GetAutonomousDatabaseTagsPerRegion", "Fetching the autonomous database resource tags from the oci")

	var fetchedResourceDetails []database.AutonomousDatabaseSummary
	var pageHeader string

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	req := database.ListAutonomousDatabasesRequest{
		CompartmentId: common.String(compartmentOCID),
	}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := od.client.ListAutonomousDatabases(od.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_database", "GetAutonomousDatabaseTagsPerRegion", err)
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
			"resource_name":   *item.DisplayName,
			"db_name":         *item.DbName,
			"db_display_name": *item.DisplayName,
			"db_version":      *item.DbVersion,
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

// GetExternalPluggableDatabaseTagsPerRegion To fetch tags from an external pluggable database, an external Oracle container database
func (od *OCIDatabase) GetExternalPluggableDatabaseTagsPerRegion(compartmentOCID string) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "GetExternalPluggableDatabaseTagsPerRegion", "Fetching the external pluggable database resource tags from the oci")

	fetchedPDBResourceDetailsChan := make(chan []database.ExternalPluggableDatabaseSummary)
	fetchedPCDResourceDetailsChan := make(chan []database.ExternalContainerDatabaseSummary)

	go od.getExternalPluggableDatabaseTags(compartmentOCID, fetchedPDBResourceDetailsChan)
	go od.getExternalContainerDatabaseTags(compartmentOCID, fetchedPCDResourceDetailsChan)

	fetchedPDBResourceDetails := <-fetchedPDBResourceDetailsChan
	fetchedPCDResourceDetails := <-fetchedPCDResourceDetailsChan

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	for _, item := range fetchedPDBResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})

		resourceLabels[*item.Id] = map[string]string{
			"resource_name":   *item.DisplayName,
			"db_unique_name":  *item.DbUniqueName,
			"db_display_name": *item.DisplayName,
			"db_version":      *item.DatabaseVersion,
		}
	}

	for _, item := range fetchedPCDResourceDetails {
		resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
			ResourceID:   *item.Id,
			ResourceName: *item.DisplayName,
			DefinedTags:  item.DefinedTags,
			FreeFormTags: item.FreeformTags,
		})

		resourceLabels[*item.Id] = map[string]string{
			"resource_name":   *item.DisplayName,
			"db_display_name": *item.DisplayName,
		}

		if item.DbUniqueName != nil {
			resourceLabels[*item.Id]["db_unique_name"] = *item.DbUniqueName
		}
		if item.DatabaseVersion != nil {
			resourceLabels[*item.Id]["db_version"] = *item.DatabaseVersion
		}
	}

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

// GetExternalContainerDatabaseTagsPerRegion To fetch tags from an external Oracle container database.
func (od *OCIDatabase) getExternalPluggableDatabaseTags(compartmentOCID string, resourceDetailsChan chan []database.ExternalPluggableDatabaseSummary) {
	backend.Logger.Debug("client.oci_database", "getExternalPluggableDatabaseTags", "Fetching the external pluggable container database resource tags from the oci")

	var fetchedResourceDetails []database.ExternalPluggableDatabaseSummary
	var pageHeader string

	req := database.ListExternalPluggableDatabasesRequest{
		CompartmentId: common.String(compartmentOCID),
	}
	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := od.client.ListExternalPluggableDatabases(od.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_database", "getExternalPluggableDatabaseTags", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	resourceDetailsChan <- fetchedResourceDetails
}

// getExternalContainerDatabaseTags To fetch tags from an external Oracle container database.
func (od *OCIDatabase) getExternalContainerDatabaseTags(compartmentOCID string, resourceDetailsChan chan []database.ExternalContainerDatabaseSummary) {
	backend.Logger.Debug("client.oci_database", "getExternalContainerDatabaseTags", "Fetching the external pluggable container database resource tags from the oci")

	var fetchedResourceDetails []database.ExternalContainerDatabaseSummary
	var pageHeader string

	req := database.ListExternalContainerDatabasesRequest{
		CompartmentId: common.String(compartmentOCID),
	}

	for {
		if len(pageHeader) != 0 {
			req.Page = common.String(pageHeader)
		}

		resp, err := od.client.ListExternalContainerDatabases(od.ctx, req)
		if err != nil {
			backend.Logger.Error("client.oci_database", "getExternalContainerDatabaseTags", err)
			break
		}

		fetchedResourceDetails = append(fetchedResourceDetails, resp.Items...)
		if len(resp.RawResponse.Header.Get("opc-next-page")) != 0 {
			pageHeader = *resp.OpcNextPage
		} else {
			break
		}
	}

	resourceDetailsChan <- fetchedResourceDetails
}
