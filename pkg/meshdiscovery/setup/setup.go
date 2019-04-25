package setup

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/registration"
)

type MeshDiscoveryOptions struct {
	DisableConfigLoop bool
}

// customCtx and customErrHandler are expected to be passed by tests
func Main(customCtx context.Context, customErrHandler func(error), opts *MeshDiscoveryOptions) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	if opts == nil {
		opts = &MeshDiscoveryOptions{}
	}

	rootCtx := createRootContext(customCtx)

	clientSet, err := clientset.ClientsetFromContext(rootCtx)
	if err != nil {
		return err
	}

	if err := mesh.RunDiscoveryEventLoop(rootCtx, clientSet, customErrHandler); err != nil {
		return err
	}

	if !opts.DisableConfigLoop {
		if err := registration.RunRegistrationEventLoop(rootCtx, clientSet, customErrHandler); err != nil {
			return err
		}
	} else {
		contextutils.LoggerFrom(rootCtx).Info("discovery config has been disabled")
	}

	<-rootCtx.Done()
	return nil
}

func createRootContext(customCtx context.Context) context.Context {
	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	rootCtx = contextutils.WithLogger(rootCtx, "meshdiscovery")
	return rootCtx
}
