package main

import (
	"fmt"
	"os"

	"github.com/solo-io/go-list-licenses/pkg/license"
)

func main() {
	glooMeshPackages := []string{
		"github.com/solo-io/gloo-mesh/cmd/cert-agent/",
		"github.com/solo-io/gloo-mesh/cmd/meshctl/",
		"github.com/solo-io/gloo-mesh/cmd/gloo-mesh/",
	}
	// dependencies for this package which are used on mac, and will not be present in linux CI
	macOnlyDependencies := []string{
		"github.com/mitchellh/go-homedir",
		"github.com/containerd/continuity",
	}
	app := license.Cli(glooMeshPackages, macOnlyDependencies)
	if err := app.Execute(); err != nil {
		fmt.Errorf("unable to run oss compliance check: %v\n", err)
		os.Exit(1)
	}
}
