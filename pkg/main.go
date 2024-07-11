/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

package main

import (
	"context"
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"

	"github.com/oracle/oci-grafana-metrics/pkg/plugin"
)

const OCI_PLUGIN_ID = "oci-metrics-datasource"

func wrappedNewOCIDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return plugin.NewOCIDatasource(settings) // forward to original function
}

func main() {
	if err := datasource.Manage(OCI_PLUGIN_ID, wrappedNewOCIDatasource, datasource.ManageOpts{}); err != nil {
		backend.Logger.Error(err.Error())
		os.Exit(1)
	}
}
