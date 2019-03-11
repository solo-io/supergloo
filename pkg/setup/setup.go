package setup

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	installsetup "github.com/solo-io/supergloo/pkg/install/istio/setup"
	registrationsetup "github.com/solo-io/supergloo/pkg/registration/setup"
)

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

	if err := registrationsetup.RunRegistrationEventLoop(rootCtx, clientSet, customErrHandler); err != nil {
		return err
	}

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
