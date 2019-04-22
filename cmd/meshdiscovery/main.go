package main

import (
	"flag"
	"log"
	"time"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/setup"
	"github.com/solo-io/supergloo/pkg/version"
)

func main() {

	if err := run(); err != nil {
		log.Fatalf("err in main: %v", err.Error())
	}
}

func run() error {
	opts := setupFlags()
	errs := make(chan error)
	start := time.Now()
	check.CallCheck("meshdiscovery", version.Version, start)
	go func() {
		errs <- setup.Main(nil, nil, opts)
	}()
	return <-errs
}

func setupFlags() *setup.MeshDiscoveryOptions {
	opts := &setup.MeshDiscoveryOptions{}
	flag.BoolVar(&opts.DisableConfigLoop, "disable-config", false, "set to true to disable mesh"+
		"config discovery")
	flag.Parse()
	return opts
}
