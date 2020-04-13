package multicluster

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	apps_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	mc_predicate "github.com/solo-io/service-mesh-hub/services/common/multicluster/predicate"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh"
	mesh_service "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-service"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/wire"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	// visible for testing
	MeshPredicates = []predicate.Predicate{
		mc_predicate.BlacklistedNamespacePredicateProvider(mc_predicate.KubeBlacklistedNamespaces),
	}
	MeshWorkloadPredicates = []predicate.Predicate{
		mc_predicate.BlacklistedNamespacePredicateProvider(
			mc_predicate.IstioBlackListedNamespaces.Union(mc_predicate.KubeBlacklistedNamespaces),
		),
	}
)

// this is the main entrypoint for discovery
// when a cluster is registered, we handle that event and spin up new resource controllers for that cluster
func NewDiscoveryClusterHandler(
	localManager mc_manager.AsyncManager,
	meshScanners []mesh.MeshScanner,
	meshWorkloadScannerFactories []mesh_workload.MeshWorkloadScannerFactory,
	discoveryContext wire.DiscoveryContext,
) (mc_manager.AsyncManagerHandler, error) {

	// these clients operate against the local cluster, so we use the local manager's client
	localClient := localManager.Manager().GetClient()
	localMeshServiceClient := discoveryContext.ClientFactories.MeshServiceClientFactory(localClient)
	localMeshWorkloadClient := discoveryContext.ClientFactories.MeshWorkloadClientFactory(localClient)
	localMeshClient := discoveryContext.ClientFactories.MeshClientFactory(localClient)

	localMeshWorkloadController, err := discoveryContext.EventWatcherFactories.MeshWorkloadEventWatcherFactory.Build(localManager, "mesh-workload-apps_controller")
	if err != nil {
		return nil, err
	}

	// we don't store the local manager on the struct to avoid mistakenly conflating the local manager with the remote manager
	handler := &discoveryClusterHandler{
		localMeshClient:               localMeshClient,
		meshScanners:                  meshScanners,
		localMeshWorkloadClient:       localMeshWorkloadClient,
		localManager:                  localManager,
		meshWorkloadScannerFactories:  meshWorkloadScannerFactories,
		discoveryContext:              discoveryContext,
		localMeshServiceClient:        localMeshServiceClient,
		localMeshWorkloadEventWatcher: localMeshWorkloadController,
	}

	return handler, nil
}

type discoveryClusterHandler struct {
	localManager     mc_manager.AsyncManager
	discoveryContext wire.DiscoveryContext

	// clients that operate against the local cluster
	localMeshClient         zephyr_discovery.MeshClient
	localMeshWorkloadClient zephyr_discovery.MeshWorkloadClient
	localMeshServiceClient  zephyr_discovery.MeshServiceClient

	// controllers that operate against the local cluster
	localMeshWorkloadEventWatcher discovery_controllers.MeshWorkloadEventWatcher

	// scanners
	meshScanners                 []mesh.MeshScanner
	meshWorkloadScannerFactories []mesh_workload.MeshWorkloadScannerFactory
}

type clusterDependentDeps struct {
	deploymentEventWatcher apps_controller.DeploymentEventWatcher
	podEventWatcher        core_controller.PodEventWatcher
	meshWorkloadScanners   []mesh_workload.MeshWorkloadScanner
	serviceEventWatcher    core_controller.ServiceEventWatcher
	serviceClient          kubernetes_core.ServiceClient
}

func (m *discoveryClusterHandler) ClusterAdded(ctx context.Context, mgr mc_manager.AsyncManager, clusterName string) error {
	initializedDeps, err := m.initializeClusterDependentDeps(mgr, clusterName)
	if err != nil {
		return err
	}
	meshFinder := mesh.NewMeshFinder(
		ctx,
		clusterName,
		m.meshScanners,
		m.localMeshClient,
		mgr.Manager().GetClient(),
	)

	meshWorkloadFinder := mesh_workload.NewMeshWorkloadFinder(
		ctx,
		clusterName,
		m.localMeshWorkloadClient,
		m.localMeshClient,
		initializedDeps.meshWorkloadScanners,
	)

	meshServiceFinder := mesh_service.NewMeshServiceFinder(
		ctx,
		clusterName,
		env.GetWriteNamespace(),
		initializedDeps.serviceClient,
		m.localMeshServiceClient,
		m.localMeshWorkloadClient,
		m.localMeshClient,
	)

	err = meshFinder.StartDiscovery(initializedDeps.deploymentEventWatcher, MeshPredicates)
	if err != nil {
		return err
	}

	err = meshWorkloadFinder.StartDiscovery(initializedDeps.podEventWatcher, MeshWorkloadPredicates)
	if err != nil {
		return err
	}

	err = meshServiceFinder.StartDiscovery(initializedDeps.serviceEventWatcher, m.localMeshWorkloadEventWatcher)
	if err != nil {
		return err
	}

	return nil
}

func (m *discoveryClusterHandler) ClusterRemoved(cluster string) error {
	// TODO: Not deleting any entities for now
	return nil
}

func (m *discoveryClusterHandler) initializeClusterDependentDeps(mgr mc_manager.AsyncManager, clusterName string) (*clusterDependentDeps, error) {
	deploymentEventWatcher, err := m.discoveryContext.EventWatcherFactories.DeploymentEventWatcherFactory.Build(mgr, clusterName)
	if err != nil {
		return nil, err
	}

	podEventWatcher, err := m.discoveryContext.EventWatcherFactories.PodEventWatcherFactory.Build(mgr, clusterName)
	if err != nil {
		return nil, err
	}

	serviceEventWatcher, err := m.discoveryContext.EventWatcherFactories.ServiceEventWatcherFactory.Build(mgr, clusterName)
	if err != nil {
		return nil, err
	}

	remoteClient := mgr.Manager().GetClient()
	var meshWorkloadScanners []mesh_workload.MeshWorkloadScanner
	for _, scannerFactory := range m.meshWorkloadScannerFactories {
		ownerFetcher := m.discoveryContext.ClientFactories.OwnerFetcherClientFactory(
			m.discoveryContext.ClientFactories.DeploymentClientFactory(remoteClient),
			m.discoveryContext.ClientFactories.ReplicaSetClientFactory(remoteClient),
		)

		meshWorkloadScanners = append(meshWorkloadScanners, scannerFactory(ownerFetcher))
	}

	serviceClient := m.discoveryContext.ClientFactories.ServiceClientFactory(remoteClient)

	return &clusterDependentDeps{
		deploymentEventWatcher: deploymentEventWatcher,
		podEventWatcher:        podEventWatcher,
		meshWorkloadScanners:   meshWorkloadScanners,
		serviceEventWatcher:    serviceEventWatcher,
		serviceClient:          serviceClient,
	}, nil
}
