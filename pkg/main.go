/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

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
