package multicluster

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/common/concurrency"
	"github.com/solo-io/mesh-projects/services/common"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	mc_predicate "github.com/solo-io/mesh-projects/services/common/multicluster/predicate"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	// visible for testing
	DeploymentPredicates = []predicate.Predicate{
		mc_predicate.BlacklistedNamespacePredicateProvider(mc_predicate.KubeBlacklistedNamespaces),
	}
)

//go:generate mockgen -destination ./mocks/mock_deployment_controller.go -package mock_multicluster github.com/solo-io/mesh-projects/services/common/cluster/apps/v1/controller DeploymentController

// this is the main entrypoint for mesh-discovery
// when a cluster is registered, we handle that event and spin up a new deployment controller for that cluster
// that new deployment controller processes Deployment events from that cluster.
// note that as a first step, the local manager should be registered so that we can detect meshes in the same
// cluster that SMH runs in. This should be handled by this constructor implementation
func NewMeshDiscoveryClusterHandler(
	ctx context.Context,
	deploymentControllerFactory DeploymentControllerFactory,
	localManager mc_manager.AsyncManager,
	localMeshClient discovery.LocalMeshClient,
	meshFinders []discovery.MeshFinder,
	safeMapBuilder func() concurrency.ThreadSafeMap,
) (mc_manager.AsyncManagerHandler, error) {

	handler := &meshDiscoveryClusterHandler{
		clusterNameToDeploymentController: safeMapBuilder(),
		deploymentControllerFactory:       deploymentControllerFactory,
		localMeshClient:                   localMeshClient,
		meshFinders:                       meshFinders,
		ctx:                               ctx,
		localManager:                      localManager,
	}

	// be sure that we are also watching our local cluster
	err := handler.ClusterAdded(localManager, common.LocalClusterName)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

type meshDiscoveryClusterHandler struct {
	clusterNameToDeploymentController concurrency.ThreadSafeMap
	deploymentControllerFactory       DeploymentControllerFactory
	localManager                      mc_manager.AsyncManager
	localMeshClient                   discovery.LocalMeshClient
	meshFinders                       []discovery.MeshFinder
	ctx                               context.Context
}

func (m *meshDiscoveryClusterHandler) ClusterAdded(mgr mc_manager.AsyncManager, clusterName string) error {
	// build a deployment controller that can receive events for the cluster that `mgr` can talk to
	deploymentController, err := m.deploymentControllerFactory.Build(mgr, clusterName)
	if err != nil {
		return err
	}

	err = deploymentController.AddEventHandler(
		m.ctx,
		discovery.NewMeshDiscoverer(m.ctx, clusterName, m.meshFinders, m.localMeshClient),
		DeploymentPredicates...,
	)
	if err != nil {
		return err
	}

	m.clusterNameToDeploymentController.Store(clusterName, deploymentController)

	return err
}

func (m *meshDiscoveryClusterHandler) ClusterRemoved(clusterName string) error {
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
