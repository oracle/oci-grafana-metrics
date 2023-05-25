package models

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	jsoniter "github.com/json-iterator/go"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin/constants"
)

// OCIDatasourceSettings holds the datasource configuration information for OCI
type OCIDatasourceSettings struct {
	AuthProvider         string `json:"authProvider"`
	ConfigPath           string `json:"configPath"`
	ConfigProfile        string `json:"configProfile"`
	MultiTenancyChoice   string `json:"multiTenancyChoice"`
	MultiTenancyMode     string `json:"multiTenancyMode"`
	MultiTenancyFile     string `json:"multiTenancyFile"`
	TenancyName          string `json:"tenancyName,omitempty"`
	EnableCMDB           bool   `json:"enableCMDB"`
	EnableCMDBUploadFile bool   `json:"enableCMDBUploadFile"`
	CMDBFileContent      string `json:"cmdbFileContent"`
}

func (d *OCIDatasourceSettings) Load(dsiSettings backend.DataSourceInstanceSettings) error {
	var err error

	if dsiSettings.JSONData != nil && len(dsiSettings.JSONData) > 1 {
		if err = jsoniter.Unmarshal(dsiSettings.JSONData, d); err != nil {
			return fmt.Errorf("could not unmarshal OCIDatasourceSettings json: %w", err)
		}
	}

	if d.AuthProvider == constants.OCI_CLI_AUTH_PROVIDER {
		homeFolder := getHomeFolder()

		if d.ConfigPath == "" || d.ConfigPath == constants.DEFAULT_CONFIG_FILE {
			d.ConfigPath = path.Join(homeFolder, constants.DEFAULT_CONFIG_DIR_NAME, "config")
		}

		if d.ConfigProfile == "" {
			d.ConfigProfile = constants.DEFAULT_PROFILE
		}

		if d.MultiTenancyChoice == constants.YES {
			if d.MultiTenancyFile == "" || d.MultiTenancyFile == constants.DEFAULT_MULTI_TENANCY_FILE {
				d.MultiTenancyFile = path.Join(homeFolder, constants.DEFAULT_CONFIG_DIR_NAME, "tenancies")
			}
		}

		if d.TenancyName == "" {
			d.TenancyName = "root"
		}

		return nil
	}

	// in case of instance principle auth provider
	d.ConfigProfile = constants.DEFAULT_INSTANCE_PROFILE

	return nil
}

func getHomeFolder() string {
	current, e := user.Current()
	if e != nil {
		//Give up and try to return something sensible
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return current.HomeDir
}
