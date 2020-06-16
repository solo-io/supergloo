package reconcilers

import (
	"context"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
	"github.com/solo-io/skv2/pkg/reconcile"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type discoveryReconcilers struct {
	ctx           context.Context
	meshDiscovery k8s.MeshDiscovery
}

func (d *discoveryReconcilers) ReconcileMeshWorkload(obj *smh_discovery.MeshWorkload) (reconcile.Result, error) {
	panic("implement me")
}

func (d *discoveryReconcilers) ReconcileMesh(obj *smh_discovery.Mesh) (reconcile.Result, error) {
	panic("implement me")
}

func (d *discoveryReconcilers) ReconcileDeployment(clusterName string, obj *apps_v1.Deployment) (reconcile.Result, error) {
	return reconcile.Result{}, d.meshDiscovery.DiscoverMesh(d.ctx, clusterName)
}

func (d *discoveryReconcilers) ReconcilePod(clusterName string, _ *v1.Pod) (reconcile.Result, error) {
	return reconcile.Result{}, d.meshDiscovery.DiscoverMesh(d.ctx, clusterName)
}

func (d *discoveryReconcilers) ReconcileService(clusterName string, obj *v1.Service) (reconcile.Result, error) {
	panic("implement me")
}
