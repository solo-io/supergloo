package mesh

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type meshReconciler struct{}

func (p *meshReconciler) ReconcileMesh(cluster string, obj *v1alpha1.Mesh) (reconcile.Result, error) {
	panic("implement me")
}
