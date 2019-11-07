package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("err in main: %v", err.Error())
	}
}

func run() error {
	errs := make(chan error)
	go func() {
		errs <- setup.Main(nil, nil)
	}()
	return <-errs
}
