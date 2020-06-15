package meshworkload

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type meshWorkloadReconciler struct{}

func (p *meshWorkloadReconciler) ReconcileMeshWorkload(cluster string, obj *v1alpha1.MeshWorkload) (reconcile.Result, error) {
	panic("implement me")
}
