package starter

import (
	"context"

	apps_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller"
	core_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"

	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/multicluster/client"
	"github.com/solo-io/skv2/pkg/reconcile"

	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/reconcilers"
	"github.com/solo-io/skv2/pkg/multicluster/watch"
	"k8s.io/apimachinery/pkg/runtime"
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

// the Starter provides the main entrypoint to run mesh-discovery.
type Starter interface {
	// Start runs an instance of mesh-discovery, using the context to block.
	// It initializes the manager.Manager for the master (local) cluster
	// and wires in the discovery reconcilers which listen/handle events in managed clusters.
	// Start will block until the Master Manager returns an error or the context is cancelled.
	Start(ctx context.Context, opts Options) error
}

type starter struct {
	// constructor for creating reconcilers
	makeReconcilers func(mcClient multicluster.Client) (reconcilers.DiscoveryReconcilers, error)
}

func (s starter) Start(ctx context.Context, opts Options) error {
	mgr, err := getMasterManager(opts)
	if err != nil {
		return err
	}

	mcWatcher, mcClient := makeMulticlusterComponents(ctx, mgr.GetScheme())

	reconcilers, err := s.makeReconcilers(mcClient)
	if err != nil {
		return err
	}

	if err := addMasterClusterReconcilers(ctx, mgr, reconcilers); err != nil {
		return err
	}

	addMultiClusterReconcilers(ctx, mcWatcher, reconcilers)

	return mgr.Start(ctx.Done())
}

// adds the reconcilers for resources watched in the master cluster
func addMasterClusterReconcilers(ctx context.Context, mgr manager.Manager, discoveryReconcilers reconcilers.DiscoveryReconcilers) error {
	err := smh_discovery_controller.
		NewMeshReconcileLoop("meshes", mgr, reconcile.Options{}).
		RunMeshReconciler(ctx, discoveryReconcilers)
	if err != nil {
		return err
	}

	err = smh_discovery_controller.
		NewMeshWorkloadReconcileLoop("mesh-workloads", mgr, reconcile.Options{}).
		RunMeshWorkloadReconciler(ctx, discoveryReconcilers)
	if err != nil {
		return err
	}

	return nil
}

// get the manager for the local cluster; we will use this as our "master" cluster
func getMasterManager(opts Options) (manager.Manager, error) {
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

// construct the handles to multi cluster functionality:
// a ClusterWatcher and a multicluster.Client
func makeMulticlusterComponents(ctx context.Context, scheme *runtime.Scheme) (multicluster.ClusterWatcher, multicluster.Client) {
	clusterWatcher := watch.NewClusterWatcher(ctx, manager.Options{
		Namespace: "", // TODO (ilackarms): support configuring specific watch namespaces on remote clusters
		Scheme:    scheme,
	})
	multiclusterClient := client.NewClient(clusterWatcher)

	return clusterWatcher, multiclusterClient
}

// adds the reconcilers for the resources
// mesh-discovery needs to watch cross-cluster
func addMultiClusterReconcilers(ctx context.Context, clusterWatcher multicluster.ClusterWatcher, discoveryReconcilers reconcilers.DiscoveryReconcilers) {

	core_v1_controller.
		NewMulticlusterServiceReconcileLoop("services", clusterWatcher).
		AddMulticlusterServiceReconciler(ctx, discoveryReconcilers)

	core_v1_controller.
		NewMulticlusterPodReconcileLoop("pods", clusterWatcher).
		AddMulticlusterPodReconciler(ctx, discoveryReconcilers)

	apps_v1_controller.
		NewMulticlusterDeploymentReconcileLoop("pods", clusterWatcher).
		AddMulticlusterDeploymentReconciler(ctx, discoveryReconcilers)

}
