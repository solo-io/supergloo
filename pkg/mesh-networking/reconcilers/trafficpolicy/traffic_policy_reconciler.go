package trafficpolicy

import (
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/reconcile"
)

type trafficPolicyReconciler struct{}

func (p *trafficPolicyReconciler) ReconcileTrafficPolicy(cluster string, obj *v1alpha1.TrafficPolicy) (reconcile.Result, error) {
	panic("implement me")
}
