package main

import (
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
	errs := make(chan error)
	start := time.Now()
	check.CallCheck("meshdiscovery", version.Version, start)
	go func() {
		errs <- setup.Main(nil, nil)
	}()
	return <-errs
}
