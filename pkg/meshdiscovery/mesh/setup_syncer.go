package mesh

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	amconfig "github.com/solo-io/supergloo/pkg/config/appmesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/appmesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/istio"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/linkerd"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func RunDiscoveryEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error)) error {
	ctx = contextutils.WithLogger(ctx, "meshdiscovery-event-loop")
	logger := contextutils.LoggerFrom(ctx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("meshdiscovery error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	plugins := configurePlugins(cs)
	meshDicoverySyncer := NewMeshDiscoverySyncer(cs.Discovery.Mesh, plugins...)

	if err := startEventLoop(ctx, errHandler, cs, meshDicoverySyncer); err != nil {
		return err
	}

	return nil
}

func configurePlugins(cs *clientset.Clientset) MeshDiscoveryPlugins {
	plugins := MeshDiscoveryPlugins{
		istio.NewIstioDiscoverySyncer(),
		linkerd.NewLinkerdDiscoverySyncer(),
		appmesh.NewAppmeshDiscoverySyncer(amconfig.NewAppMeshClientBuilder(cs.Input.Secret), cs.Input.Secret),
	}
	return plugins
}

// start the mesh discovery event loop
func startEventLoop(ctx context.Context, errHandler func(err error), c *clientset.Clientset, syncer v1.DiscoverySyncer) error {
	meshDiscoveryEmitter := v1.NewDiscoverySimpleEmitter(wrapper.AggregatedWatchFromClients(
		wrapper.ClientWatchOpts{
			BaseClient: c.Input.Pod.BaseClient(),
		},
		wrapper.ClientWatchOpts{
			BaseClient: c.Input.Install.BaseClient(),
		},
		// only need to watch for the appmesh configmap
		wrapper.ClientWatchOpts{
			BaseClient:   c.Input.ConfigMap.BaseClient(),
			Namespace:    appmesh.AwsConfigMapNamespace,
			ResourceName: appmesh.AwsConfigMapName,
		},
	))

	meshDiscoveryEventLoop := v1.NewDiscoverySimpleEventLoop(meshDiscoveryEmitter, syncer)

	meshDiscoveryErrs, err := meshDiscoveryEventLoop.Run(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-meshDiscoveryErrs:
				errHandler(err)
			case <-ctx.Done():
			}
		}
	}()
	return nil
}
