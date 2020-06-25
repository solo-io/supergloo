package reconciler

import (
	"context"

	apps_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller"
	core_v1_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/reconcile"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// the discovery reconciler reconciles events for watched resources
// by performing a global discovery sync
type DiscoveryReconciler interface {
	core_v1_controller.MulticlusterPodReconciler
	core_v1_controller.MulticlusterServiceReconciler
	core_v1_controller.MulticlusterConfigMapReconciler
	apps_v1_controller.MulticlusterDeploymentReconciler
	apps_v1_controller.MulticlusterReplicaSetReconciler
	apps_v1_controller.MulticlusterDaemonSetReconciler
	apps_v1_controller.MulticlusterStatefulSetReconciler
}

var _ DiscoveryReconciler = &discoveryReconciler{}

type discoveryReconciler struct {
	ctx          context.Context
	builder      input.Builder
	translator   translation.Translator
	masterClient client.Client
}

func Start(
	ctx context.Context,
	builder input.Builder,
	translator translation.Translator,
	masterClient client.Client,
	clusters multicluster.ClusterWatcher,
) {
	d := &discoveryReconciler{
		ctx:          ctx,
		builder:      builder,
		translator:   translator,
		masterClient: masterClient,
	}

	core_v1_controller.NewMulticlusterPodReconcileLoop("Pod", clusters).AddMulticlusterPodReconciler(ctx, d)
	core_v1_controller.NewMulticlusterServiceReconcileLoop("Service", clusters).AddMulticlusterServiceReconciler(ctx, d)
	core_v1_controller.NewMulticlusterConfigMapReconcileLoop("ConfigMap", clusters).AddMulticlusterConfigMapReconciler(ctx, d)
	apps_v1_controller.NewMulticlusterDeploymentReconcileLoop("Deployment", clusters).AddMulticlusterDeploymentReconciler(ctx, d)
	apps_v1_controller.NewMulticlusterReplicaSetReconcileLoop("ReplicaSet", clusters).AddMulticlusterReplicaSetReconciler(ctx, d)
	apps_v1_controller.NewMulticlusterDaemonSetReconcileLoop("DaemonSet", clusters).AddMulticlusterDaemonSetReconciler(ctx, d)
	apps_v1_controller.NewMulticlusterStatefulSetReconcileLoop("StatefulSet", clusters).AddMulticlusterStatefulSetReconciler(ctx, d)

}

func (d *discoveryReconciler) ReconcilePod(clusterName string, obj *corev1.Pod) (reconcile.Result, error) {
	contextutils.LoggerFrom(d.ctx).Debugw("reconciling event", "cluster", clusterName, "obj", obj)
	return reconcile.Result{}, d.reconcile()
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
		// failed to read from cache; should never happen
		return err
	}

	outputSnap, err := d.translator.Translate(d.ctx, inputSnap)
	if err != nil {
		// internal translator errors should never happen
		return err
	}

	return outputSnap.Apply(d.ctx, d.masterClient)
}
