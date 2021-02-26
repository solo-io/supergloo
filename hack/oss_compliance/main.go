package main

import (
	"fmt"
	"os"

	"github.com/solo-io/go-list-licenses/pkg/license"
)

func main() {
	macOnlyDependencies := []string{
		"github.com/mitchellh/go-homedir",
		"github.com/containerd/continuity",
	}
	app, err := license.CliAllPackages(macOnlyDependencies)
	if err != nil {
		fmt.Println("unable to gather all gloo mesh packages")
		os.Exit(1)
	}
	if err := app.Execute(); err != nil {
		fmt.Printf("unable to run oss compliance check: %v\n", err)
		os.Exit(1)
	}
}
