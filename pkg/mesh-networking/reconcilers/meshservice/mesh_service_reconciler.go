package meshservice

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type meshServiceReconciler struct{}

func (p *meshServiceReconciler) ReconcileMeshService(cluster string, obj *v1alpha1.MeshService) (reconcile.Result, error) {
	panic("implement me")
}
