package virtualmesh

import (
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type virtualMeshReconciler struct{}

func (p *virtualMeshReconciler) ReconcileVirtualMesh(cluster string, obj *v1alpha1.VirtualMesh) (reconcile.Result, error) {
	panic("implement me")
}
