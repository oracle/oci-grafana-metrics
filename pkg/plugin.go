// Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	log.DefaultLogger.Debug("Running GRPC server")

	if err := datasource.Manage("myorgid-simple-backend-datasource", NewOCIDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error("Errore nel plugingo")
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
	log.DefaultLogger.Error("plugingo OK")
}
