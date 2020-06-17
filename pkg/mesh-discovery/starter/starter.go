package starter

import (
	"context"

	k8s_apps_controller "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller"
	k8s_apps_providers "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/providers"
	k8s_core_clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	k8s_core_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s/linkerd"

	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/multicluster/client"
	"github.com/solo-io/skv2/pkg/reconcile"

	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/reconcilers"
	"github.com/solo-io/skv2/pkg/multicluster/watch"
	"k8s.io/apimachinery/pkg/runtime"
	controller_runtime "sigs.k8s.io/controller-runtime/pkg/client"
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

type starter struct{}

func NewStarter() Starter {
	return &starter{}
}

func (s starter) Start(ctx context.Context, opts Options) error {
	mgr, err := getMasterManager(opts)
	if err != nil {
		return err
	}

	mcWatcher, mcClient := makeMulticlusterComponents(ctx, mgr.GetScheme())

	reconcilers := s.initializeReconcilers(ctx, mgr.GetClient(), mcClient)

	if err := addMasterClusterReconcilers(ctx, mgr, reconcilers); err != nil {
		return err
	}

	addMultiClusterReconcilers(ctx, mcWatcher, reconcilers)

	return mgr.Start(ctx.Done())
}

// constructor for creating reconcilers
func (s starter) initializeReconcilers(
	ctx context.Context,
	masterClient controller_runtime.Client,
	mcClient multicluster.Client,
) reconcilers.DiscoveryReconcilers {
	meshClient := smh_discovery.NewMeshClient(masterClient)
	meshWorkloadClient := smh_discovery.NewMeshWorkloadClient(masterClient)
	meshWorkloadScanners := initializeMeshWorkloadScanners(mcClient, meshClient)
	meshWorkloadDiscovery := meshworkload_discovery.NewMeshWorkloadDiscovery(
		meshClient,
		meshWorkloadClient,
		meshWorkloadScanners,
		k8s_core_clients.NewMulticlusterClientset(mcClient),
	)
	return reconcilers.NewDiscoveryReconcilers(ctx, meshWorkloadDiscovery)
}

func initializeMeshWorkloadScanners(
	mcClient multicluster.Client,
	meshClient smh_discovery.MeshClient,
) meshworkload_discovery.MeshWorkloadScanners {
	ownerFetcherFactory := meshworkload_discovery.NewOwnerFetcherFactory(
		k8s_apps_providers.DeploymentClientFactoryProvider(),
		k8s_apps_providers.ReplicaSetClientFactoryProvider(),
		mcClient,
	)
	arnParser := aws_utils.NewArnParser()
	appmeshScanner := aws_utils.NewAppMeshScanner(arnParser)
	configMapClientFactory := k8s_core_providers.ConfigMapClientFactoryProvider()
	awsAccountIdFetcher := aws_utils.NewAwsAccountIdFetcher(arnParser, configMapClientFactory, mcClient)
	appmeshWorkloadScanner := appmesh.NewAppMeshWorkloadScanner(
		ownerFetcherFactory,
		appmeshScanner,
		meshClient,
		awsAccountIdFetcher,
	)
	istioMeshWorkloadScanner := istio.NewIstioMeshWorkloadScanner(ownerFetcherFactory, meshClient)
	linkerdMeshWorkloadScanner := linkerd.NewLinkerdMeshWorkloadScanner(ownerFetcherFactory, meshClient)
	scannerFactories := meshworkload_discovery.MeshWorkloadScanners{
		types.MeshType_ISTIO1_5: istioMeshWorkloadScanner,
		types.MeshType_ISTIO1_6: istioMeshWorkloadScanner,
		types.MeshType_LINKERD:  linkerdMeshWorkloadScanner,
		types.MeshType_APPMESH:  appmeshWorkloadScanner,
	}
	return scannerFactories
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

	k8s_core_controller.
		NewMulticlusterServiceReconcileLoop("services", clusterWatcher).
		AddMulticlusterServiceReconciler(ctx, discoveryReconcilers)

	k8s_core_controller.
		NewMulticlusterPodReconcileLoop("pods", clusterWatcher).
		AddMulticlusterPodReconciler(ctx, discoveryReconcilers)

	k8s_apps_controller.
		NewMulticlusterDeploymentReconcileLoop("pods", clusterWatcher).
		AddMulticlusterDeploymentReconciler(ctx, discoveryReconcilers)

}
