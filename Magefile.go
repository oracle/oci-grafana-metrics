// Copyright © 2022 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

//go:build mage

package main

import (
	"fmt"
	"os"

	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
)

// Default configures the default target.
var Default = build.BuildAll

// Cleans up local folder
func CleanLocal() error {
	fmt.Println("Cleans the local folder")
	err := os.RemoveAll("oci-metrics-datasource/")
	if err != nil {
		fmt.Println(err)
	}
	return err

}
