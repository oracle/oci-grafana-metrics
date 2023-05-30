package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin"
)

const OCI_PLUGIN_ID = "oci-metrics-datasource"

func main() {
	if err := datasource.Manage(OCI_PLUGIN_ID, plugin.NewOCIDatasource, datasource.ManageOpts{}); err != nil {
		backend.Logger.Error(err.Error())
		os.Exit(1)
	}
}
