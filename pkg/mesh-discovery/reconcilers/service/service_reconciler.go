package service

import (
	"github.com/solo-io/skv2/pkg/reconcile"
	v1 "k8s.io/api/core/v1"
)

type serviceReconciler struct{}

func (p *serviceReconciler) ReconcileService(cluster string, obj *v1.Service) (reconcile.Result, error) {
	panic("implement me")
}
