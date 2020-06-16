package service

import (
	"github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	meshservice_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	"github.com/solo-io/skv2/pkg/reconcile"
	v1 "k8s.io/api/core/v1"
)

type serviceReconciler struct {
	meshServiceFinder meshservice_discovery.MeshServiceFinder
}

func NewServiceReconciler(meshServiceFinder meshservice_discovery.MeshServiceFinder) controller.MulticlusterServiceReconciler {
	return &serviceReconciler{meshServiceFinder: meshServiceFinder}
}

func (s *serviceReconciler) ReconcileService(cluster string, _ *v1.Service) (reconcile.Result, error) {
	return reconcile.Result{}, s.meshServiceFinder.Reconcile(cluster)
}
