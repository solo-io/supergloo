package reconciler

import (
	"context"
	apps_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller"
	core_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/pkg/reconcile"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/input"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/output"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// the discovery reconciler reconciles events for watched resources
// by performing a global discovery sync
type DiscoveryReconcilers interface {
	core_v1_controller.MulticlusterPodReconciler
	core_v1_controller.MulticlusterServiceReconciler
	core_v1_controller.MulticlusterConfigMapReconciler
	apps_v1_controller.MulticlusterDeploymentReconciler
	apps_v1_controller.MulticlusterReplicaSetReconciler
	apps_v1_controller.MulticlusterDaemonSetReconciler
	apps_v1_controller.MulticlusterStatefulSetReconciler
}

type discoveryReconciler struct {
	ctx        context.Context
	builder    input.Builder
	translator translation.Translator
	applier    output.Applier
}


func (d *discoveryReconciler) ReconcileService(clusterName string, obj *corev1.Service) (reconcile.Result, error) {
	contextutils.LoggerFrom(d.ctx).Debugw("reconciling event", "cluster", clusterName, "obj", obj)
	return reconcile.Result{}, d.reconcile()
}

func (d *discoveryReconciler) ReconcileConfigMap(clusterName string, obj *corev1.ConfigMap) (reconcile.Result, error) {
	contextutils.LoggerFrom(d.ctx).Debugw("reconciling event", "cluster", clusterName, "obj", obj)
	return reconcile.Result{}, d.reconcile()
}

func (d *discoveryReconciler) ReconcileDeployment(clusterName string, obj *appsv1.Deployment) (reconcile.Result, error) {
	contextutils.LoggerFrom(d.ctx).Debugw("reconciling event", "cluster", clusterName, "obj", obj)
	return reconcile.Result{}, d.reconcile()
}

func (d *discoveryReconciler) ReconcileReplicaSet(clusterName string, obj *appsv1.ReplicaSet) (reconcile.Result, error) {
	contextutils.LoggerFrom(d.ctx).Debugw("reconciling event", "cluster", clusterName, "obj", obj)
	return reconcile.Result{}, d.reconcile()
}

func (d *discoveryReconciler) ReconcileDaemonSet(clusterName string, obj *appsv1.DaemonSet) (reconcile.Result, error) {
	contextutils.LoggerFrom(d.ctx).Debugw("reconciling event", "cluster", clusterName, "obj", obj)
	return reconcile.Result{}, d.reconcile()
}

func (d *discoveryReconciler) ReconcileStatefulSet(clusterName string, obj *appsv1.StatefulSet) (reconcile.Result, error) {
	contextutils.LoggerFrom(d.ctx).Debugw("reconciling event", "cluster", clusterName, "obj", obj)
	return reconcile.Result{}, d.reconcile()
}

// reconcile global state
func (d *discoveryReconciler) reconcile() error {
	inputSnap, err := d.builder.BuildSnapshot(d.ctx)
	if err != nil {
		return err
	}

	outputSnap := d.translator.Translate(inputSnap)

	return d.applier.Apply(outputSnap)
}
