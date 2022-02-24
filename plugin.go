// Copyright © 2019 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package main

import (
		"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
		"github.com/grafana/grafana-plugin-sdk-go/backend/log"
		"os"
)

func main() {
		log.DefaultLogger.Debug("Running GRPC server")

		if err := datasource.Manage("myorgid-simple-backend-datasource", NewOCIDatasource, datasource.ManageOpts{}); err != nil {
				log.DefaultLogger.Error(err.Error())
				os.Exit(1)
		}
}
