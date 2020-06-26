package mesh_discovery

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/multicluster/watch"
	"github.com/solo-io/smh/pkg/mesh-discovery/reconciler"
	"github.com/solo-io/smh/pkg/mesh-discovery/translator"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// bootstrap options for starting discovery
// TODO: wire these up to Settings CR
type Options struct {
	// MetricsBindAddress is the TCP address that the controller should bind to
	// for serving prometheus metrics.
	// It can be set to "0" to disable the metrics serving.
	MetricsBindAddress string

	// MasterNamespace if specified restricts the Master manager's cache to watch objects in
	// the desired namespace Defaults to all namespaces
	//
	// Note: If a namespace is specified, controllers can still Watch for a
	// cluster-scoped resource (e.g Node).  For namespaced resources the cache
	// will only hold objects from the desired namespace.
	MasterNamespace string
}

// the mesh-discovery controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts Options) error {
	mgr, err := makeMasterManager(opts)
	if err != nil {
		return err
	}

	clusterWatcher := watch.NewClusterWatcher(ctx, manager.Options{
		Namespace: "", // TODO (ilackarms): support configuring specific watch namespaces on remote clusters
		Scheme:    mgr.GetScheme(),
	})

	mcClient := multicluster.NewClient(clusterWatcher)

	startReconciler(ctx, mgr.GetClient(), mcClient, clusterWatcher, clusterWatcher)

	return mgr.Start(ctx.Done())
}

// get the manager for the local cluster; we will use this as our "master" cluster
func makeMasterManager(opts Options) (manager.Manager, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          opts.MasterNamespace, // TODO (ilackarms): support configuring multiple watch namespaces on master cluster
		MetricsBindAddress: opts.MetricsBindAddress,
	})
	if err != nil {
		return nil, err
	}

	if err := schemes.SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}
	return mgr, nil
}

// constructor for creating reconcilers
func startReconciler(
	ctx context.Context,
	masterClient client.Client,
	mcClient multicluster.Client,
	clusters multicluster.ClusterSet,
	mcWatcher multicluster.ClusterWatcher,
) {
	snapshotBuilder := input.NewMultiClusterBuilder(clusters, mcClient)
	translator := translator.NewTranslator()
	reconciler.Start(ctx, snapshotBuilder, translator, masterClient, mcWatcher)
}
