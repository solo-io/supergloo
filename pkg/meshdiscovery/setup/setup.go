package setup

import (
	"context"
	"os"

	"github.com/solo-io/supergloo/pkg/meshdiscovery/linkerd"

	"github.com/solo-io/supergloo/pkg/meshdiscovery/istio"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
)

// customCtx and customErrHandler are expected to be passed by tests
func Main(customCtx context.Context, errHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	writeNamespace := os.Getenv("POD_NAMESPACE")
	if writeNamespace == "" {
		writeNamespace = "supergloo-system"
	}

	rootCtx := createRootContext(customCtx)

	clientSet, err := clientset.ClientsetFromContext(rootCtx)
	if err != nil {
		return err
	}

	istioClients, err := clientset.IstioClientsetFromContext(rootCtx)
	if err != nil {
		return err
	}

	if err := runDiscoveryEventLoop(rootCtx, writeNamespace, clientSet, istioClients, errHandler); err != nil {
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
	rootCtx = contextutils.WithLogger(rootCtx, "meshdiscovery")
	return rootCtx
}

func runDiscoveryEventLoop(ctx context.Context, writeNamespace string, cs *clientset.Clientset, istioClients *clientset.IstioClientset, errHandler func(error)) error {

	meshReconciler := v1.NewMeshReconciler(cs.Discovery.Mesh)

	istioDiscovery := istio.NewIstioDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		istioClients.MeshPolicies,
		cs.ApiExtensions.ApiextensionsV1beta1().CustomResourceDefinitions(),
		cs.Kube.BatchV1(),
	)

	linkerdDiscovery := linkerd.NewLinkerdDiscoverySyncer(
		writeNamespace,
		meshReconciler,
	)

	emitter := v1.NewDiscoverySimpleEmitter(wrapper.AggregatedWatchFromClients(
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Deployment.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Upstream.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Pod.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.ConfigMap.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.TlsSecret.BaseClient()},
	))
	eventLoop := v1.NewDiscoverySimpleEventLoop(emitter,
		istioDiscovery,
		linkerdDiscovery,
	)

	errs, err := eventLoop.Run(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-errs:
				errHandler(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
