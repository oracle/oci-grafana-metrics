//go:build mage
// +build mage

package main

import (
	"fmt"
	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
)

// Hello prints a message (shows that you can define custom Mage targets).
func Hello() {
	fmt.Println("hello! We are building binary only for linux amd64!")
}

// Default configures the default target.
var Default = build.BuildAll
