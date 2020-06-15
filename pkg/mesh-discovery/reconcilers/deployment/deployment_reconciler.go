package deployment

import (
	"github.com/solo-io/skv2/pkg/reconcile"
	"k8s.io/api/apps/v1"
)

type deploymentReconciler struct{}

func (p *deploymentReconciler) ReconcileDeployment(cluster string, obj *v1.Deployment) (reconcile.Result, error) {
	panic("implement me")
}
