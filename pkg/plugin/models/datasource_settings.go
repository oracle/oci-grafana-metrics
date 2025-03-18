/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

package models

import (
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	jsoniter "github.com/json-iterator/go"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
)

// OCIDatasourceSettings holds the datasource configuration information for OCI
type OCIDatasourceSettings struct {
	AuthProvider  string `json:"authProvider"`
	ConfigProfile string `json:"configProfile"`
	TenancyMode   string `json:"tenancymode"`
	TenancyName   string `json:"tenancyName,omitempty"`
	Environment   string `json:"environment"`

	Profile_0 string `json:"profile0,omitempty"`
	Region_0  string `json:"region0,omitempty"`

	Profile_1 string `json:"profile1,omitempty"`
	Region_1  string `json:"region1,omitempty"`

	Profile_2 string `json:"profile2,omitempty"`
	Region_2  string `json:"region2,omitempty"`

	Profile_3 string `json:"profile3,omitempty"`
	Region_3  string `json:"region3,omitempty"`

	Profile_4 string `json:"profile4,omitempty"`
	Region_4  string `json:"region4,omitempty"`

	Profile_5 string `json:"profile5,omitempty"`
	Region_5  string `json:"region5,omitempty"`

	Xtenancy_0 string `json:"xtenancy0,omitempty"`

	CustomRegion_0 string `json:"customregion0,omitempty"`
	CustomRegion_1 string `json:"customregion1,omitempty"`
	CustomRegion_2 string `json:"customregion2,omitempty"`
	CustomRegion_3 string `json:"customregion3,omitempty"`
	CustomRegion_4 string `json:"customregion4,omitempty"`
	CustomRegion_5 string `json:"customregion5,omitempty"`
}

// Load initializes the OCIDatasourceSettings from the provided backend.DataSourceInstanceSettings.
// It unmarshals the JSONData from the DataSourceInstanceSettings into the OCIDatasourceSettings struct.
// If the JSONData is not nil and has more than one element, it attempts to unmarshal it.
// If unmarshalling fails, it returns an error indicating the failure.
// Additionally, it sets the ConfigProfile to the default instance profile.
//
// Parameters:
// - dsiSettings: backend.DataSourceInstanceSettings containing the settings to load.
//
// Returns:
// - error: An error if the JSONData could not be unmarshalled, otherwise nil.
func (d *OCIDatasourceSettings) Load(dsiSettings backend.DataSourceInstanceSettings) error {
	var err error

	if dsiSettings.JSONData != nil && len(dsiSettings.JSONData) > 1 {
		if err = jsoniter.Unmarshal(dsiSettings.JSONData, d); err != nil {
			return fmt.Errorf("could not unmarshal OCIDatasourceSettings json: %w", err)
		}
	}

	// in case of instance principle auth provider
	d.ConfigProfile = constants.DEFAULT_INSTANCE_PROFILE

	return nil
}
