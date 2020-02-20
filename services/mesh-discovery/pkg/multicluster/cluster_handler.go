package multicluster

import (
	"context"

	controller4 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common"
	"github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	controller2 "github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
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
// when a cluster is registered, we handle that event and spin up a new deployment controller for that cluster
// that new deployment controller processes Deployment events from that cluster.
// note that as a first step, the local manager should be registered so that we can detect meshes in the same
// cluster that SMH runs in. This should be handled by this constructor implementation
func NewDiscoveryClusterHandler(
	ctx context.Context,
	localManager mc_manager.AsyncManager,
	localMeshClient zephyr_core.MeshClient,
	meshScanners []mesh.MeshScanner,
	meshWorkloadScannerFactories []mesh_workload.MeshWorkloadScannerFactory,
	localMeshWorkloadClient zephyr_core.MeshWorkloadClient,
	discoveryContext wire.DiscoveryContext,
) (mc_manager.AsyncManagerHandler, error) {

	handler := &discoveryClusterHandler{
		localMeshClient:              localMeshClient,
		meshScanners:                 meshScanners,
		localMeshWorkloadClient:      localMeshWorkloadClient,
		ctx:                          ctx,
		localManager:                 localManager,
		meshWorkloadScannerFactories: meshWorkloadScannerFactories,
		discoveryContext:             discoveryContext,
	}

	// be sure that we are also watching our local cluster
	err := handler.ClusterAdded(localManager, common.LocalClusterName)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

type discoveryClusterHandler struct {
	ctx              context.Context
	localManager     mc_manager.AsyncManager
	discoveryContext wire.DiscoveryContext

	// clients that are instantiated already
	localMeshClient         zephyr_core.MeshClient
	localMeshWorkloadClient zephyr_core.MeshWorkloadClient

	// scanners
	meshScanners                 []mesh.MeshScanner
	meshWorkloadScannerFactories []mesh_workload.MeshWorkloadScannerFactory
}

type clusterDependentDeps struct {
	deploymentController   controller.DeploymentController
	podController          controller2.PodController
	meshWorkloadScanners   []mesh_workload.MeshWorkloadScanner
	serviceController      controller2.ServiceController
	meshWorkloadController controller4.MeshWorkloadController
	serviceClient          kubernetes_core.ServiceClient
	meshWorkloadClient     zephyr_core.MeshWorkloadClient
	meshServiceClient      zephyr_core.MeshServiceClient
}

func (m *discoveryClusterHandler) ClusterAdded(mgr mc_manager.AsyncManager, clusterName string) error {
	initializedDeps, err := m.initializeClusterDependentDeps(mgr, clusterName)
	if err != nil {
		return err
	}

	meshFinder := mesh.NewMeshFinder(
		m.ctx,
		clusterName,
		m.meshScanners,
		m.localMeshClient,
	)

	meshWorkloadFinder := mesh_workload.NewMeshWorkloadFinder(
		m.ctx,
		clusterName,
		m.localMeshWorkloadClient,
		m.localMeshClient,
		initializedDeps.meshWorkloadScanners,
	)

	meshServiceFinder := mesh_service.NewMeshServiceFinder(
		m.ctx,
		clusterName,
		env.DefaultWriteNamespace,
		initializedDeps.serviceClient,
		initializedDeps.meshServiceClient,
		initializedDeps.meshWorkloadClient,
	)

	err = meshFinder.StartDiscovery(initializedDeps.deploymentController, MeshPredicates)
	if err != nil {
		return err
	}

	err = meshWorkloadFinder.StartDiscovery(initializedDeps.podController, MeshWorkloadPredicates)
	if err != nil {
		return err
	}

	err = meshServiceFinder.StartDiscovery(initializedDeps.serviceController, initializedDeps.meshWorkloadController)
	if err != nil {
		return err
	}

	return nil
}

func (m *discoveryClusterHandler) ClusterRemoved(clusterName string) error {
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

	// ensure that the mesh workload client is watching our local cluster- use the local manager
	meshWorkloadController, err := m.discoveryContext.ControllerFactories.MeshWorkloadControllerFactory.Build(m.localManager, clusterName)
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

	// these clients operate against the local cluster, so we use the local manager's client
	localClient := m.localManager.Manager().GetClient()
	meshServiceClient := m.discoveryContext.ClientFactories.MeshServiceClientFactory(localClient)
	meshWorkloadClient := m.discoveryContext.ClientFactories.MeshWorkloadClientFactory(localClient)

	return &clusterDependentDeps{
		deploymentController:   deploymentController,
		podController:          podController,
		meshWorkloadScanners:   meshWorkloadScanners,
		serviceController:      serviceController,
		meshWorkloadController: meshWorkloadController,
		serviceClient:          serviceClient,
		meshWorkloadClient:     meshWorkloadClient,
		meshServiceClient:      meshServiceClient,
	}, nil
}
