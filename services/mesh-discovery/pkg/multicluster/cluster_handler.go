package multicluster

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_apps_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh"
	mesh_service "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-service"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/wire"
)

type MeshWorkloadScannerFactoryImplementations map[core_types.MeshType]mesh_workload.MeshWorkloadScannerFactory

// this is the main entrypoint for discovery
// when a cluster is registered, we handle that event and spin up new resource controllers for that cluster
func NewDiscoveryClusterHandler(
	localManager mc_manager.AsyncManager,
	meshScanners []mesh.MeshScanner,
	meshWorkloadScannerFactories MeshWorkloadScannerFactoryImplementations,
	discoveryContext wire.DiscoveryContext,
) (mc_manager.AsyncManagerHandler, error) {

	// these clients operate against the local cluster, so we use the local manager's client
	localClient := localManager.Manager().GetClient()
	localMeshServiceClient := discoveryContext.ClientFactories.MeshServiceClientFactory(localClient)
	localMeshWorkloadClient := discoveryContext.ClientFactories.MeshWorkloadClientFactory(localClient)
	localMeshClient := discoveryContext.ClientFactories.MeshClientFactory(localClient)

	localMeshWorkloadEventWatcher := discoveryContext.EventWatcherFactories.MeshWorkloadEventWatcherFactory.Build(localManager, "mesh-workload-apps_controller")

	localMeshController := discoveryContext.EventWatcherFactories.MeshControllerFactory.Build(localManager, "mesh-controller")

	// we don't store the local manager on the struct to avoid mistakenly conflating the local manager with the remote manager
	handler := &discoveryClusterHandler{
		localMeshClient:               localMeshClient,
		meshScanners:                  meshScanners,
		localMeshWorkloadClient:       localMeshWorkloadClient,
		localManager:                  localManager,
		meshWorkloadScannerFactories:  meshWorkloadScannerFactories,
		discoveryContext:              discoveryContext,
		localMeshServiceClient:        localMeshServiceClient,
		localMeshWorkloadEventWatcher: localMeshWorkloadEventWatcher,
		localMeshEventWatcher:         localMeshController,
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
	localMeshWorkloadEventWatcher zephyr_discovery_controller.MeshWorkloadEventWatcher
	localMeshEventWatcher         zephyr_discovery_controller.MeshEventWatcher

	// scanners
	meshScanners                 []mesh.MeshScanner
	meshWorkloadScannerFactories MeshWorkloadScannerFactoryImplementations
}

type clusterDependentDeps struct {
	deploymentEventWatcher k8s_apps_controller.DeploymentEventWatcher
	podEventWatcher        k8s_core_controller.PodEventWatcher
	meshWorkloadScanners   mesh_workload.MeshWorkloadScannerImplementations
	serviceEventWatcher    k8s_core_controller.ServiceEventWatcher
	serviceClient          k8s_core.ServiceClient
	podClient              k8s_core.PodClient
	deploymentClient       k8s_apps.DeploymentClient
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
		initializedDeps.deploymentClient,
	)

	meshWorkloadFinder := mesh_workload.NewMeshWorkloadFinder(
		ctx,
		clusterName,
		m.localMeshWorkloadClient,
		m.localMeshClient,
		initializedDeps.meshWorkloadScanners,
		initializedDeps.podClient,
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

	err = meshFinder.StartDiscovery(initializedDeps.deploymentEventWatcher)
	if err != nil {
		return err
	}

	err = meshWorkloadFinder.StartDiscovery(initializedDeps.podEventWatcher, m.localMeshEventWatcher)
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
	deploymentEventWatcher := m.discoveryContext.EventWatcherFactories.DeploymentEventWatcherFactory.Build(mgr, clusterName)
	podEventWatcher := m.discoveryContext.EventWatcherFactories.PodEventWatcherFactory.Build(mgr, clusterName)
	serviceEventWatcher := m.discoveryContext.EventWatcherFactories.ServiceEventWatcherFactory.Build(mgr, clusterName)

	remoteClient := mgr.Manager().GetClient()

	serviceClient := m.discoveryContext.ClientFactories.ServiceClientFactory(remoteClient)
	podClient := m.discoveryContext.ClientFactories.PodClientFactory(remoteClient)
	deploymentClient := m.discoveryContext.ClientFactories.DeploymentClientFactory(remoteClient)
	replicaSetClient := m.discoveryContext.ClientFactories.ReplicaSetClientFactory(remoteClient)

	meshWorkloadScanners := make(mesh_workload.MeshWorkloadScannerImplementations)
	for meshType, scannerFactory := range m.meshWorkloadScannerFactories {
		ownerFetcher := m.discoveryContext.ClientFactories.OwnerFetcherClientFactory(
			deploymentClient,
			replicaSetClient,
		)

		meshWorkloadScanners[meshType] = scannerFactory(ownerFetcher)
	}

	return &clusterDependentDeps{
		deploymentEventWatcher: deploymentEventWatcher,
		podEventWatcher:        podEventWatcher,
		meshWorkloadScanners:   meshWorkloadScanners,
		serviceEventWatcher:    serviceEventWatcher,
		serviceClient:          serviceClient,
		podClient:              podClient,
		deploymentClient:       deploymentClient,
	}, nil
}
