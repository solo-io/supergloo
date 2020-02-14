package multicluster

import (
	"context"

	"github.com/hashicorp/go-multierror"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/common/concurrency"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/solo-io/mesh-projects/services/common"
	"github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	mc_predicate "github.com/solo-io/mesh-projects/services/common/multicluster/predicate"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh"
	mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh-workload"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/controllers"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	// visible for testing
	ObjectPredicates = []predicate.Predicate{
		mc_predicate.BlacklistedNamespacePredicateProvider(mc_predicate.KubeBlacklistedNamespaces),
	}
)

// this is the main entrypoint for discovery
// when a cluster is registered, we handle that event and spin up a new deployment controller for that cluster
// that new deployment controller processes Deployment events from that cluster.
// note that as a first step, the local manager should be registered so that we can detect meshes in the same
// cluster that SMH runs in. This should be handled by this constructor implementation
func NewDiscoveryClusterHandler(
	ctx context.Context,
	imageParser docker.ImageNameParser,
	localManager mc_manager.AsyncManager,
	deploymentControllerFactory controllers.DeploymentControllerFactory,
	localMeshClient zephyr_core.MeshClient,
	meshScanners []mesh.MeshScanner,
	podControllerFactory controllers.PodControllerFactory,
	localMeshWorkloadClient zephyr_core.MeshWorkloadClient,
	safeMapBuilder func() concurrency.ThreadSafeMap,
) (mc_manager.AsyncManagerHandler, error) {

	handler := &discoveryClusterHandler{
		clusterNameToDeploymentController: safeMapBuilder(),
		deploymentControllerFactory:       deploymentControllerFactory,
		localMeshClient:                   localMeshClient,
		meshScanners:                      meshScanners,
		podControllerFactory:              podControllerFactory,
		localMeshWorkloadClient:           localMeshWorkloadClient,
		ctx:                               ctx,
		localManager:                      localManager,
		imageParser:                       imageParser,
	}

	// be sure that we are also watching our local cluster
	err := handler.ClusterAdded(localManager, common.LocalClusterName)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

type discoveryClusterHandler struct {
	clusterNameToDeploymentController concurrency.ThreadSafeMap
	localManager                      mc_manager.AsyncManager
	deploymentControllerFactory       controllers.DeploymentControllerFactory
	localMeshClient                   zephyr_core.MeshClient
	meshScanners                      []mesh.MeshScanner
	podControllerFactory              controllers.PodControllerFactory
	localMeshWorkloadClient           zephyr_core.MeshWorkloadClient
	eventHandlers                     []*controller.DeploymentEventHandler
	ctx                               context.Context
	imageParser                       docker.ImageNameParser
}

func (m *discoveryClusterHandler) ClusterAdded(mgr mc_manager.AsyncManager, clusterName string) error {
	var multiErr *multierror.Error
	err := m.initializeDeploymentController(mgr, clusterName)
	if err != nil {
		multierror.Append(multiErr, err)
	}
	err = m.initializePodController(mgr, clusterName)
	if err != nil {
		multierror.Append(multiErr, err)
	}
	return err
}

func (m *discoveryClusterHandler) ClusterRemoved(clusterName string) error {
	// TODO: Not deleting any entities for now
	m.clusterNameToDeploymentController.Delete(clusterName)
	return nil
	//mesh, err := m.localMeshClient.Get(m.ctx, client.ObjectKey{
	//	Namespace: env.DefaultWriteNamespace,
	//	Name:      name,
	//})
	//
	//if err != nil {
	//	return err
	//}
	//
	//err = m.localMeshClient.Delete(m.ctx, mesh)
	//if err != nil {
	//	return err
	//}
	//
	//m.clusterNameToDeploymentController.Delete(name)
	//return nil
}

func (m *discoveryClusterHandler) initializeDeploymentController(mgr mc_manager.AsyncManager, clusterName string) error {
	// build a deployment controller that can receive events for the cluster that `mgr` can talk to
	deploymentController, err := m.deploymentControllerFactory.Build(mgr, clusterName)
	if err != nil {
		return err
	}
	err = deploymentController.AddEventHandler(
		m.ctx,
		mesh.DefaultMeshFinder(m.ctx, clusterName, m.meshScanners, m.localMeshClient),
		ObjectPredicates...,
	)
	if err != nil {
		return err
	}
	m.clusterNameToDeploymentController.Store(clusterName, deploymentController)
	return nil
}

func (m *discoveryClusterHandler) initializePodController(mgr mc_manager.AsyncManager, clusterName string) error {
	// build a deployment controller that can receive events for the cluster that `mgr` can talk to
	podController, err := m.podControllerFactory.Build(mgr, clusterName)
	if err != nil {
		return err
	}
	meshWorkloadFinder := mesh_workload.DefaultMeshWorkloadFinder(
		clusterName,
		m.ctx,
		m.localMeshWorkloadClient,
		m.localMeshClient,
		mgr.Manager().GetClient(),
		m.imageParser,
	)
	err = podController.AddEventHandler(
		m.ctx,
		meshWorkloadFinder,
		ObjectPredicates...,
	)
	if err != nil {
		return err
	}
	return nil
}
