package meshworkload

import (
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	meshservice_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type meshWorkloadReconciler struct {
	meshServiceFinder meshservice_discovery.MeshServiceFinder
}

func NewMeshWorkloadReconciler(meshServiceFinder meshservice_discovery.MeshServiceFinder) controller.MeshWorkloadReconciler {
	return &meshWorkloadReconciler{meshServiceFinder: meshServiceFinder}
}

func (m *meshWorkloadReconciler) ReconcileMeshWorkload(obj *smh_discovery.MeshWorkload) (reconcile.Result, error) {
	cluster := obj.Spec.GetKubeController().GetKubeControllerRef().GetCluster()
	return reconcile.Result{}, m.meshServiceFinder.Reconcile(cluster)
}
