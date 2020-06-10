package compute_target

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_apps_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1/controller"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	meshservice_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/wire"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
)

type MeshWorkloadScannerFactoryImplementations map[core_types.MeshType]meshworkload_discovery.MeshWorkloadScannerFactory

// this is the main entrypoint for discovery
// when a cluster is registered, we handle that event and spin up new resource controllers for that cluster
func NewDiscoveryClusterHandler(
	localManager mc_manager.AsyncManager,
	meshScanners []k8s.MeshScanner,
	meshWorkloadScannerFactories MeshWorkloadScannerFactoryImplementations,
	clusterTenancyScannerFactories []k8s_tenancy.ClusterTenancyScannerFactory,
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
		localMeshClient:                localMeshClient,
		meshScanners:                   meshScanners,
		localMeshWorkloadClient:        localMeshWorkloadClient,
		localManager:                   localManager,
		meshWorkloadScannerFactories:   meshWorkloadScannerFactories,
		discoveryContext:               discoveryContext,
		localMeshServiceClient:         localMeshServiceClient,
		localMeshWorkloadEventWatcher:  localMeshWorkloadEventWatcher,
		localMeshEventWatcher:          localMeshController,
		clusterTenancyScannerFactories: clusterTenancyScannerFactories,
	}

	return handler, nil
}

type discoveryClusterHandler struct {
	localManager     mc_manager.AsyncManager
	discoveryContext wire.DiscoveryContext

	// clients that operate against the local cluster
	localMeshClient         smh_discovery.MeshClient
	localMeshWorkloadClient smh_discovery.MeshWorkloadClient
	localMeshServiceClient  smh_discovery.MeshServiceClient

	// controllers that operate against the local cluster
	localMeshWorkloadEventWatcher smh_discovery_controller.MeshWorkloadEventWatcher
	localMeshEventWatcher         smh_discovery_controller.MeshEventWatcher

	// scanners
	meshScanners                   []k8s.MeshScanner
	meshWorkloadScannerFactories   MeshWorkloadScannerFactoryImplementations
	clusterTenancyScannerFactories []k8s_tenancy.ClusterTenancyScannerFactory
}

type clusterDependentDeps struct {
	deploymentEventWatcher k8s_apps_controller.DeploymentEventWatcher
	podEventWatcher        k8s_core_controller.PodEventWatcher
	meshWorkloadScanners   meshworkload_discovery.MeshWorkloadScannerImplementations
	clusterTenancyScanners []k8s_tenancy.ClusterTenancyRegistrar
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
	meshFinder := k8s.NewMeshFinder(
		ctx,
		clusterName,
		m.meshScanners,
		m.localMeshClient,
		mgr.Manager().GetClient(),
		initializedDeps.deploymentClient,
	)

	meshWorkloadFinder := meshworkload_discovery.NewMeshWorkloadFinder(
		ctx,
		clusterName,
		m.localMeshClient,
		m.localMeshWorkloadClient,
		initializedDeps.meshWorkloadScanners,
		initializedDeps.podClient,
	)

	meshServiceFinder := meshservice_discovery.NewMeshServiceFinder(
		ctx,
		clusterName,
		container_runtime.GetWriteNamespace(),
		initializedDeps.serviceClient,
		m.localMeshServiceClient,
		m.localMeshWorkloadClient,
		m.localMeshClient,
	)

	clusterTenancyFinder := k8s_tenancy.NewClusterTenancyFinder(
		clusterName,
		initializedDeps.clusterTenancyScanners,
		initializedDeps.podClient,
		m.localMeshClient,
	)

	if err = meshFinder.StartDiscovery(initializedDeps.deploymentEventWatcher); err != nil {
		return err
	}

	if err = clusterTenancyFinder.StartRegistration(ctx, initializedDeps.podEventWatcher, m.localMeshEventWatcher); err != nil {
		return err
	}

	if err = meshWorkloadFinder.StartDiscovery(initializedDeps.podEventWatcher, m.localMeshEventWatcher); err != nil {
		return err
	}

	return meshServiceFinder.StartDiscovery(initializedDeps.serviceEventWatcher, m.localMeshWorkloadEventWatcher)
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

	meshWorkloadScanners := make(meshworkload_discovery.MeshWorkloadScannerImplementations)
	for meshType, scannerFactory := range m.meshWorkloadScannerFactories {
		ownerFetcher := m.discoveryContext.ClientFactories.OwnerFetcherClientFactory(
			deploymentClient,
			replicaSetClient,
		)
		meshWorkloadScanners[meshType] = scannerFactory(ownerFetcher, m.localMeshClient, mgr.Manager().GetClient())
	}

	var clusterTenancyScanners []k8s_tenancy.ClusterTenancyRegistrar
	for _, tenancyScannerFactory := range m.clusterTenancyScannerFactories {
		clusterTenancyScanners = append(clusterTenancyScanners, tenancyScannerFactory(m.localMeshClient, mgr.Manager().GetClient()))
	}

	return &clusterDependentDeps{
		deploymentEventWatcher: deploymentEventWatcher,
		podEventWatcher:        podEventWatcher,
		meshWorkloadScanners:   meshWorkloadScanners,
		serviceEventWatcher:    serviceEventWatcher,
		serviceClient:          serviceClient,
		podClient:              podClient,
		deploymentClient:       deploymentClient,
		clusterTenancyScanners: clusterTenancyScanners,
	}, nil
}
