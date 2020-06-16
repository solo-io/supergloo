package reconcilers

import (
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	meshservice_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	"github.com/solo-io/skv2/pkg/reconcile"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type discoveryReconcilers struct {
	meshServiceFinder meshservice_discovery.MeshServiceFinder
}

func (d *discoveryReconcilers) ReconcileMeshWorkload(obj *smh_discovery.MeshWorkload) (reconcile.Result, error) {
	clusterName := obj.Spec.GetKubeController().GetKubeControllerRef().GetCluster()
	return reconcile.Result{}, d.meshServiceFinder.Reconcile(clusterName)
}

func (d *discoveryReconcilers) ReconcileMesh(obj *smh_discovery.Mesh) (reconcile.Result, error) {
	panic("implement me")
}

func (d *discoveryReconcilers) ReconcileDeployment(clusterName string, obj *apps_v1.Deployment) (reconcile.Result, error) {
	panic("implement me")
}

func (d *discoveryReconcilers) ReconcilePod(clusterName string, obj *v1.Pod) (reconcile.Result, error) {
	panic("implement me")
}

func (d *discoveryReconcilers) ReconcileService(clusterName string, obj *v1.Service) (reconcile.Result, error) {
	return reconcile.Result{}, d.meshServiceFinder.Reconcile(clusterName)
}
