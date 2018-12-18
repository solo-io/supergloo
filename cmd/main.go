package main

import (
	"flag"
	"log"
	"strings"
	"time"

	check "github.com/solo-io/go-checkpoint"
	"github.com/solo-io/supergloo/pkg/setup"
	"github.com/solo-io/supergloo/pkg/version"
)

// TODO (ilackarms): move to a flags package
type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
func main() {
	if err := run(); err != nil {
		log.Fatalf("err in main: %v", err.Error())
	}
}

func run() error {
	errs := make(chan error)
	start := time.Now()
	check.CallCheck("supergloo", version.Version, start)
	go func() {
		var namespaces arrayFlags
		flag.Var(&namespaces, "n", "namespace to watch for crds")
		flag.Parse()
		errs <- setup.Main(nil, namespaces...)
	}()
	return <-errs
}
