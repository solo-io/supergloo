package setup

import (
	"context"
	"flag"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	"github.com/solo-io/supergloo/pkg/config/setup"
	installsetup "github.com/solo-io/supergloo/pkg/install/setup"
	"github.com/solo-io/supergloo/pkg/registration"
	registrationsetup "github.com/solo-io/supergloo/pkg/registration/setup"
)

type Options struct {
	Registration registrationsetup.RegistrationOptions
}

// customCtx and customErrHandler are expected to be passed by tests
func Main(customCtx context.Context, customErrHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	rootCtx := createRootContext(customCtx)

	clientSet, err := clientset.ClientsetFromContext(rootCtx)
	if err != nil {
		return err
	}

	if err := installsetup.RunInstallEventLoop(rootCtx, clientSet, customErrHandler); err != nil {
		return err
	}

	pubsub := registration.NewPubsub()

	opts := optsFromFlags()

	if err := registrationsetup.RunRegistrationEventLoop(rootCtx, clientSet, customErrHandler, pubsub, opts.Registration); err != nil {
		return err
	}

	setup.StartSuperglooConfigLoop(rootCtx, clientSet, pubsub)

	<-rootCtx.Done()
	return nil
}

func createRootContext(customCtx context.Context) context.Context {
	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	rootCtx = contextutils.WithLogger(rootCtx, "supergloo")
	return rootCtx
}

func optsFromFlags() Options {
	var opts Options

	flag.BoolVar(&opts.Registration.DisablePrometheusBouncer, "disable-prometheus-bouncer", false, "disable automatic "+
		"bouncing of prometheus pods on config reload. enable tihs option if using the prometheus configmap reload")
	flag.Parse()

	return opts
}
