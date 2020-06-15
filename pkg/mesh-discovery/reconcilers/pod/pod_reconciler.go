package pod

import (
	"github.com/solo-io/skv2/pkg/reconcile"
	v1 "k8s.io/api/core/v1"
)

type podReconciler struct{}

func (p *podReconciler) ReconcilePod(cluster string, obj *v1.Pod) (reconcile.Result, error) {
	panic("implement me")
}
