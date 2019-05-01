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
	var version, imageTag string
	if len(os.Args) < 3 {
		panic("Must provide version as argument")
	} else {
		version, imageTag = os.Args[1], os.Args[2]
	}
	log.Printf("Generating helm files")
	if err := generate.Run(version, imageTag, alwaysPull, ""); err != nil {
		panic(err)
	}
}
