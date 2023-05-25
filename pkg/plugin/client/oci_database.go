package client

import (
	"context"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/database"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/models"
)

type OCIDatabase struct {
	ctx    context.Context
	client database.DatabaseClient
}

// getDatabaseHomes to fetch db home details
func (od *OCIDatabase) getDatabaseHomes(compartment models.OCIResource) []map[string]string {
	backend.Logger.Debug("client.oci_database", "getDatabaseHomes", "Fetching the database homes from the oci for compartment: "+compartment.Name)

	var fetchedResourceDetails []database.DbHomeSummary
	var pageHeader string

	resourceInfo := []map[string]string{}

	req := database.ListDbHomesRequest{
		CompartmentId: common.String(compartment.OCID),
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

// getOracleDatabaseTagsPerCompartment To fetch tags from an Oracle Database on a bare metal or virtual machine DB system per compartment
func (od *OCIDatabase) getOracleDatabaseTagsPerCompartment(compartment models.OCIResource) (map[string]map[string]struct{}, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "getOracleDatabaseTagsPerCompartment", "Fetching the database resource tags from the oci for compartment: "+compartment.Name)

	var fetchedResourceDetails []database.DatabaseSummary
	var pageHeader string

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	// fetching the db homes
	dbHomes := od.getDatabaseHomes(compartment)

	for _, dbHome := range dbHomes {
		req := database.ListDatabasesRequest{
			CompartmentId: common.String(compartment.OCID),
			DbHomeId:      common.String(dbHome["db_home_id"]),
		}

		for {
			if len(pageHeader) != 0 {
				req.Page = common.String(pageHeader)
			}

			resp, err := od.client.ListDatabases(od.ctx, req)
			if err != nil {
				backend.Logger.Error("client.oci_database", "getOracleDatabaseTagsPerCompartment", err)
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
				"compartment":    compartment.Name,
				"db_name":        *item.DbName,
				"db_unique_name": *item.DbUniqueName,
				"db_home_name":   dbHome["db_home_name"],
				"db_version":     dbHome["db_version"],
			}

			if item.PdbName != nil {
				resourceLabels[*item.Id]["pdb_name"] = *item.PdbName
			}
		}
	}

	resourceTags, resourceIDsPerTag := collectResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

// GetOracleDatabaseTagsPerCompartment To fetch tags from an Oracle Database on a bare metal or virtual machine DB system per region
func (od *OCIDatabase) GetOracleDatabaseTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "GetOracleDatabaseTagsPerRegion", "Fetching the database resource tags from the oci")

	// when queried for a single compartment
	if len(compartments) == 1 {
		resourceTags, resourceIDsPerTag, apmLabels := od.getOracleDatabaseTagsPerCompartment(compartments[0])
		return convertToArray(resourceTags), resourceIDsPerTag, apmLabels
	}

	// holds key: value1, value2, for UI
	allResourceTags := map[string]map[string]struct{}{}
	// holds key.value: map of resourceIDs, for caching
	allResourceIDsPerTag := map[string]map[string]struct{}{}
	allResourceLabels := map[string]map[string]string{}

	var allCompartmentOracleDatabaseData sync.Map
	var wg sync.WaitGroup

	// fetching data per compartment
	for _, compartmentInAction := range compartments {
		if compartmentInAction.OCID == "" {
			continue
		}

		wg.Add(1)

		go func(compartment models.OCIResource) {
			defer wg.Done()

			resourceTags, resourceIDsPerTag, resourceLabels := od.getOracleDatabaseTagsPerCompartment(compartment)

			allCompartmentOracleDatabaseData.Store(compartment.OCID, map[string]interface{}{
				"resourceTags":      resourceTags,
				"resourceIDsPerTag": resourceIDsPerTag,
				"resourceLabels":    resourceLabels,
			})

		}(compartmentInAction)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentOracleDatabaseData.Range(func(key, value interface{}) bool {
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

// GetAutonomousDatabaseTagsPerRegion To fetch tags from an Oracle Autonomous Database.
func (od *OCIDatabase) GetAutonomousDatabaseTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "GetAutonomousDatabaseTagsPerRegion", "Fetching the autonomous database resource tags from the oci")

	resourceLabels := map[string]map[string]string{}
	resourceTagsResponse := []models.OCIResourceTagsResponse{}

	var pageHeader string
	var allCompartmentData sync.Map
	var wg sync.WaitGroup

	// fetching data per compartment
	for _, compartmentInAction := range compartments {
		if compartmentInAction.OCID == "" {
			continue
		}

		wg.Add(1)

		go func(resource models.OCIResource) {
			defer wg.Done()

			var fetchedResourceDetails []database.AutonomousDatabaseSummary

			req := database.ListAutonomousDatabasesRequest{
				CompartmentId: common.String(resource.OCID),
			}

			for {
				if len(pageHeader) != 0 {
					req.Page = common.String(pageHeader)
				}

				resp, err := od.client.ListAutonomousDatabases(od.ctx, req)
				if err != nil {
					backend.Logger.Warn("client.oci_database", "GetAutonomousDatabaseTagsPerRegion:REQ", req)
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

			allCompartmentData.Store(resource.Name, fetchedResourceDetails)
		}(compartmentInAction)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentData.Range(func(key, value interface{}) bool {
		compartmentName := key.(string)
		fetchedResourceData := value.([]database.AutonomousDatabaseSummary)

		for _, item := range fetchedResourceData {
			resourceTagsResponse = append(resourceTagsResponse, models.OCIResourceTagsResponse{
				ResourceID:   *item.Id,
				ResourceName: *item.DisplayName,
				DefinedTags:  item.DefinedTags,
				FreeFormTags: item.FreeformTags,
			})

			resourceLabels[*item.Id] = map[string]string{
				"resource_name":   *item.DisplayName,
				"compartment":     compartmentName,
				"db_name":         *item.DbName,
				"db_display_name": *item.DisplayName,
				"db_version":      *item.DbVersion,
			}
		}

		return true
	})

	resourceTags, resourceIDsPerTag := fetchResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

// GetExternalContainerDatabaseTagsPerRegion To fetch tags from an external Oracle container database.
func (od *OCIDatabase) getExternalPluggableDatabaseTags(compartment models.OCIResource, resourceDetailsChan chan []database.ExternalPluggableDatabaseSummary) {
	backend.Logger.Debug("client.oci_database", "getExternalPluggableDatabaseTags", "Fetching the external pluggable container database resource tags from the oci compartment: "+compartment.Name)

	var fetchedResourceDetails []database.ExternalPluggableDatabaseSummary
	var pageHeader string

	req := database.ListExternalPluggableDatabasesRequest{
		CompartmentId: common.String(compartment.OCID),
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
func (od *OCIDatabase) getExternalContainerDatabaseTags(compartment models.OCIResource, resourceDetailsChan chan []database.ExternalContainerDatabaseSummary) {
	backend.Logger.Debug("client.oci_database", "getExternalContainerDatabaseTags", "Fetching the external pluggable container database resource tags from the oci for compartment: "+compartment.Name)

	var fetchedResourceDetails []database.ExternalContainerDatabaseSummary
	var pageHeader string

	req := database.ListExternalContainerDatabasesRequest{
		CompartmentId: common.String(compartment.OCID),
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

// getExternalDatabaseTagsPerCompartment To fetch tags from an external pluggable database, an external Oracle container database per compartment
func (od *OCIDatabase) getExternalDatabaseTagsPerCompartment(compartment models.OCIResource) (map[string]map[string]struct{}, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "getExternalDatabaseTagsPerCompartment", "Fetching the external pluggable database resource tags from the oci for compartment: "+compartment.Name)

	fetchedPDBResourceDetailsChan := make(chan []database.ExternalPluggableDatabaseSummary)
	fetchedPCDResourceDetailsChan := make(chan []database.ExternalContainerDatabaseSummary)

	go od.getExternalPluggableDatabaseTags(compartment, fetchedPDBResourceDetailsChan)
	go od.getExternalContainerDatabaseTags(compartment, fetchedPCDResourceDetailsChan)

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
			"compartment":     compartment.Name,
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

	resourceTags, resourceIDsPerTag := collectResourceTags(resourceTagsResponse)

	return resourceTags, resourceIDsPerTag, resourceLabels
}

// GetExternalPluggableDatabaseTagsPerRegion To fetch tags from an external pluggable database, an external Oracle container database per region
func (od *OCIDatabase) GetExternalPluggableDatabaseTagsPerRegion(compartments []models.OCIResource) (map[string][]string, map[string]map[string]struct{}, map[string]map[string]string) {
	backend.Logger.Debug("client.oci_database", "GetExternalPluggableDatabaseTagsPerRegion", "Fetching the external pluggable database resource tags from the oci")

	// when queried for a single compartment
	if len(compartments) == 1 {
		resourceTags, resourceIDsPerTag, apmLabels := od.getExternalDatabaseTagsPerCompartment(compartments[0])
		return convertToArray(resourceTags), resourceIDsPerTag, apmLabels
	}

	// holds key: value1, value2, for UI
	allResourceTags := map[string]map[string]struct{}{}
	// holds key.value: map of resourceIDs, for caching
	allResourceIDsPerTag := map[string]map[string]struct{}{}
	allResourceLabels := map[string]map[string]string{}

	var allCompartmentExternalPluggableDatabaseData sync.Map
	var wg sync.WaitGroup

	// fetching data per compartment
	for _, compartmentInAction := range compartments {
		if compartmentInAction.OCID == "" {
			continue
		}

		wg.Add(1)

		go func(compartment models.OCIResource) {
			defer wg.Done()

			resourceTags, resourceIDsPerTag, resourceLabels := od.getExternalDatabaseTagsPerCompartment(compartment)

			allCompartmentExternalPluggableDatabaseData.Store(compartment.OCID, map[string]interface{}{
				"resourceTags":      resourceTags,
				"resourceIDsPerTag": resourceIDsPerTag,
				"resourceLabels":    resourceLabels,
			})

		}(compartmentInAction)
	}
	wg.Wait()

	// collecting the data from all compartments
	allCompartmentExternalPluggableDatabaseData.Range(func(key, value interface{}) bool {
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
