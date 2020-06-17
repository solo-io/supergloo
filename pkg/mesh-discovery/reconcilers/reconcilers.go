package reconcilers

import (
	"context"

	apps_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller"
	core_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	meshservice_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	"github.com/solo-io/skv2/pkg/reconcile"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type discoveryReconcilers struct {
	ctx                   context.Context
	meshServiceDiscovery  meshservice_discovery.MeshServiceDiscovery
	meshWorkloadDiscovery meshworkload_discovery.MeshWorkloadDiscovery
}

func NewDiscoveryReconcilers(
	ctx context.Context,
	meshWorkloadDiscovery meshworkload_discovery.MeshWorkloadDiscovery,
	meshServiceDiscovery meshservice_discovery.MeshServiceDiscovery,
) DiscoveryReconcilers {
	return &discoveryReconcilers{
		ctx:                   ctx,
		meshServiceDiscovery:  meshServiceDiscovery,
		meshWorkloadDiscovery: meshWorkloadDiscovery,
	}
}

type DiscoveryReconcilers interface {
	smh_discovery_controller.MeshWorkloadReconciler
	smh_discovery_controller.MeshReconciler

	apps_v1_controller.MulticlusterDeploymentReconciler
	core_v1_controller.MulticlusterPodReconciler
	core_v1_controller.MulticlusterServiceReconciler
}

func (d *discoveryReconcilers) ReconcileMeshWorkload(obj *smh_discovery.MeshWorkload) (reconcile.Result, error) {
	clusterName := obj.Spec.GetKubeController().GetKubeControllerRef().GetCluster()
	return reconcile.Result{}, d.meshServiceDiscovery.DiscoverMeshServices(d.ctx, clusterName)
}

func (d *discoveryReconcilers) ReconcileMesh(obj *smh_discovery.Mesh) (reconcile.Result, error) {
	clusterName := obj.Spec.GetCluster().GetName()
	return reconcile.Result{}, d.meshWorkloadDiscovery.DiscoverMeshWorkloads(d.ctx, clusterName)
}

func (d *discoveryReconcilers) ReconcileDeployment(clusterName string, obj *apps_v1.Deployment) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (d *discoveryReconcilers) ReconcilePod(clusterName string, obj *v1.Pod) (reconcile.Result, error) {
	return reconcile.Result{}, d.meshWorkloadDiscovery.DiscoverMeshWorkloads(d.ctx, clusterName)
}

func (d *discoveryReconcilers) ReconcileService(clusterName string, obj *v1.Service) (reconcile.Result, error) {
	return reconcile.Result{}, d.meshServiceDiscovery.DiscoverMeshServices(d.ctx, clusterName)
}
