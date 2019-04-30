package setup

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/config/appmesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/config/istio"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/config/linkerd"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh"
	discreg "github.com/solo-io/supergloo/pkg/meshdiscovery/registration"
	"github.com/solo-io/supergloo/pkg/registration"
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
		pubsub := registration.NewPubsub()
		if err := discreg.RunRegistrationEventLoop(rootCtx, clientSet, customErrHandler, pubsub); err != nil {
			return err
		}

		newDiscoveryConfigLoops(rootCtx, clientSet, pubsub)
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

func newDiscoveryConfigLoops(ctx context.Context, clientset *clientset.Clientset, pubsub *registration.PubSub) {
	istio.StartIstioDiscoveryConfigLoop(ctx, clientset, pubsub)
	linkerd.StartLinkerdDiscoveryConfigLoop(ctx, clientset, pubsub)
	appmesh.StartAppmeshDiscoveryConfigLoop(ctx, clientset, pubsub)
}
