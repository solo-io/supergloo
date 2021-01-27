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

	app := license.Cli(glooMeshPackages)
	if err := app.Execute(); err != nil {
		fmt.Errorf("unable to run oss compliance check: %v\n", err)
		os.Exit(1)
	}
}
