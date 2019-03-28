package main

import (
	"os"

	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/supergloo/install/helm/supergloo/generate"
)

const (
	neverPull    = "Never"
	alwaysPull   = "Always"
	ifNotPresent = "IfNotPresent"
)

func main() {
	var version string
	if len(os.Args) < 2 {
		panic("Must provide version as argument")
	} else {
		version = os.Args[1]
	}
	log.Printf("Generating helm files.")
	pullPolicy := alwaysPull
	if version == "dev" {
		pullPolicy = neverPull
	}
	if err := generate.Run(version, pullPolicy); err != nil {
		panic(err)
	}
}
