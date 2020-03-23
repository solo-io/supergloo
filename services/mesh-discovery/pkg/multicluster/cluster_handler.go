package multicluster

import (
	"context"

	discovery_controllers "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	corev1_controllers "github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	mc_predicate "github.com/solo-io/mesh-projects/services/common/multicluster/predicate"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh"
	mesh_service "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-service"
	mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-workload"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/wire"
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

	localMeshWorkloadController, err := discoveryContext.ControllerFactories.MeshWorkloadControllerFactory.Build(localManager, "mesh-workload-controller")
	if err != nil {
		return nil, err
	}

	// we don't store the local manager on the struct to avoid mistakenly conflating the local manager with the remote manager
	handler := &discoveryClusterHandler{
		localMeshClient:              localMeshClient,
		meshScanners:                 meshScanners,
		localMeshWorkloadClient:      localMeshWorkloadClient,
		localManager:                 localManager,
		meshWorkloadScannerFactories: meshWorkloadScannerFactories,
		discoveryContext:             discoveryContext,
		localMeshServiceClient:       localMeshServiceClient,
		localMeshWorkloadController:  localMeshWorkloadController,
	}

	return handler, nil
}

type discoveryClusterHandler struct {
	localManager     mc_manager.AsyncManager
	discoveryContext wire.DiscoveryContext

	// clients that operate against the local cluster
	localMeshClient         zephyr_core.MeshClient
	localMeshWorkloadClient zephyr_core.MeshWorkloadClient
	localMeshServiceClient  zephyr_core.MeshServiceClient

	// controllers that operate against the local cluster
	localMeshWorkloadController discovery_controllers.MeshWorkloadController

	// scanners
	meshScanners                 []mesh.MeshScanner
	meshWorkloadScannerFactories []mesh_workload.MeshWorkloadScannerFactory
}

type clusterDependentDeps struct {
	deploymentController controller.DeploymentController
	podController        corev1_controllers.PodController
	meshWorkloadScanners []mesh_workload.MeshWorkloadScanner
	serviceController    corev1_controllers.ServiceController
	serviceClient        kubernetes_core.ServiceClient
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
		env.DefaultWriteNamespace,
		initializedDeps.serviceClient,
		m.localMeshServiceClient,
		m.localMeshWorkloadClient,
	)

	err = meshFinder.StartDiscovery(initializedDeps.deploymentController, MeshPredicates)
	if err != nil {
		return err
	}

	err = meshWorkloadFinder.StartDiscovery(initializedDeps.podController, MeshWorkloadPredicates)
	if err != nil {
		return err
	}

	err = meshServiceFinder.StartDiscovery(initializedDeps.serviceController, m.localMeshWorkloadController)
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
	deploymentController, err := m.discoveryContext.ControllerFactories.DeploymentControllerFactory.Build(mgr, clusterName)
	if err != nil {
		return nil, err
	}

	podController, err := m.discoveryContext.ControllerFactories.PodControllerFactory.Build(mgr, clusterName)
	if err != nil {
		return nil, err
	}

	serviceController, err := m.discoveryContext.ControllerFactories.ServiceControllerFactory.Build(mgr, clusterName)
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
		deploymentController: deploymentController,
		podController:        podController,
		meshWorkloadScanners: meshWorkloadScanners,
		serviceController:    serviceController,
		serviceClient:        serviceClient,
	}, nil
}
