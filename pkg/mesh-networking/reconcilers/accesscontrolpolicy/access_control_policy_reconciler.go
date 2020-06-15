package accesscontrolpolicy

import (
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type accessControlPolicyReconciler struct{}

func (p *accessControlPolicyReconciler) ReconcileAccessControlPolicy(cluster string, obj *v1alpha1.AccessControlPolicy) (reconcile.Result, error) {
	panic("implement me")
}
