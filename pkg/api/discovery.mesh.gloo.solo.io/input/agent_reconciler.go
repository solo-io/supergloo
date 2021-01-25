// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./agent_reconciler.go -destination mocks/agent_reconciler.go

// The Input Reconciler calls a simple func() error whenever a
// storage event is received for any of:
// * SettingsMeshGlooSoloIov1Alpha2Settings
// * AppmeshK8SAwsv1Beta2Meshes
// * V1ConfigMaps
// * V1Services
// * V1Pods
// * V1Nodes
// * Appsv1Deployments
// * Appsv1ReplicaSets
// * Appsv1DaemonSets
// * Appsv1StatefulSets
// for a given cluster or set of clusters.
//
// Input Reconcilers can be be constructed from either a single Manager (watch events in a single cluster)
// or a ClusterWatcher (watch events in multiple clusters).
package input

import (
	"context"
	"time"

	"github.com/solo-io/skv2/contrib/pkg/input"
	sk_core_v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	settings_mesh_gloo_solo_io_v1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	settings_mesh_gloo_solo_io_v1alpha2_controllers "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2/controller"

	appmesh_k8s_aws_v1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	appmesh_k8s_aws_v1beta2_controllers "github.com/solo-io/external-apis/pkg/api/appmesh/appmesh.k8s.aws/v1beta2/controller"

	v1_controllers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	v1 "k8s.io/api/core/v1"

	apps_v1_controllers "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/controller"
	apps_v1 "k8s.io/api/apps/v1"
)

// the multiClusterAgentReconciler reconciles events for input resources across clusters
// this private interface is used to ensure that the generated struct implements the intended functions
type multiClusterAgentReconciler interface {
	settings_mesh_gloo_solo_io_v1alpha2_controllers.MulticlusterSettingsReconciler

	appmesh_k8s_aws_v1beta2_controllers.MulticlusterMeshReconciler

	v1_controllers.MulticlusterConfigMapReconciler
	v1_controllers.MulticlusterServiceReconciler
	v1_controllers.MulticlusterPodReconciler
	v1_controllers.MulticlusterNodeReconciler

	apps_v1_controllers.MulticlusterDeploymentReconciler
	apps_v1_controllers.MulticlusterReplicaSetReconciler
	apps_v1_controllers.MulticlusterDaemonSetReconciler
	apps_v1_controllers.MulticlusterStatefulSetReconciler
}

var _ multiClusterAgentReconciler = &multiClusterAgentReconcilerImpl{}

type multiClusterAgentReconcilerImpl struct {
	base input.InputReconciler
}

// Options for reconciling a snapshot
type AgentReconcileOptions struct {

	// Options for reconciling SettingsMeshGlooSoloIov1Alpha2Settings
	SettingsMeshGlooSoloIov1Alpha2Settings reconcile.Options

	// Options for reconciling AppmeshK8SAwsv1Beta2Meshes
	AppmeshK8SAwsv1Beta2Meshes reconcile.Options

	// Options for reconciling V1ConfigMaps
	V1ConfigMaps reconcile.Options
	// Options for reconciling V1Services
	V1Services reconcile.Options
	// Options for reconciling V1Pods
	V1Pods reconcile.Options
	// Options for reconciling V1Nodes
	V1Nodes reconcile.Options

	// Options for reconciling Appsv1Deployments
	Appsv1Deployments reconcile.Options
	// Options for reconciling Appsv1ReplicaSets
	Appsv1ReplicaSets reconcile.Options
	// Options for reconciling Appsv1DaemonSets
	Appsv1DaemonSets reconcile.Options
	// Options for reconciling Appsv1StatefulSets
	Appsv1StatefulSets reconcile.Options
}

// register the reconcile func with the cluster watcher
// the reconcileInterval, if greater than 0, will limit the number of reconciles
// to one per interval.
func RegisterMultiClusterAgentReconciler(
	ctx context.Context,
	clusters multicluster.ClusterWatcher,
	reconcileFunc input.MultiClusterReconcileFunc,
	reconcileInterval time.Duration,
	options AgentReconcileOptions,
	predicates ...predicate.Predicate,
) input.InputReconciler {

	base := input.NewInputReconciler(
		ctx,
		reconcileFunc,
		nil,
		reconcileInterval,
	)

	r := &multiClusterAgentReconcilerImpl{
		base: base,
	}

	// initialize reconcile loops

	settings_mesh_gloo_solo_io_v1alpha2_controllers.NewMulticlusterSettingsReconcileLoop("Settings", clusters, options.SettingsMeshGlooSoloIov1Alpha2Settings).AddMulticlusterSettingsReconciler(ctx, r, predicates...)

	appmesh_k8s_aws_v1beta2_controllers.NewMulticlusterMeshReconcileLoop("Mesh", clusters, options.AppmeshK8SAwsv1Beta2Meshes).AddMulticlusterMeshReconciler(ctx, r, predicates...)

	v1_controllers.NewMulticlusterConfigMapReconcileLoop("ConfigMap", clusters, options.V1ConfigMaps).AddMulticlusterConfigMapReconciler(ctx, r, predicates...)

	v1_controllers.NewMulticlusterServiceReconcileLoop("Service", clusters, options.V1Services).AddMulticlusterServiceReconciler(ctx, r, predicates...)

	v1_controllers.NewMulticlusterPodReconcileLoop("Pod", clusters, options.V1Pods).AddMulticlusterPodReconciler(ctx, r, predicates...)

	v1_controllers.NewMulticlusterNodeReconcileLoop("Node", clusters, options.V1Nodes).AddMulticlusterNodeReconciler(ctx, r, predicates...)

	apps_v1_controllers.NewMulticlusterDeploymentReconcileLoop("Deployment", clusters, options.Appsv1Deployments).AddMulticlusterDeploymentReconciler(ctx, r, predicates...)

	apps_v1_controllers.NewMulticlusterReplicaSetReconcileLoop("ReplicaSet", clusters, options.Appsv1ReplicaSets).AddMulticlusterReplicaSetReconciler(ctx, r, predicates...)

	apps_v1_controllers.NewMulticlusterDaemonSetReconcileLoop("DaemonSet", clusters, options.Appsv1DaemonSets).AddMulticlusterDaemonSetReconciler(ctx, r, predicates...)

	apps_v1_controllers.NewMulticlusterStatefulSetReconcileLoop("StatefulSet", clusters, options.Appsv1StatefulSets).AddMulticlusterStatefulSetReconciler(ctx, r, predicates...)
	return r.base
}

func (r *multiClusterAgentReconcilerImpl) ReconcileSettings(clusterName string, obj *settings_mesh_gloo_solo_io_v1alpha2.Settings) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileSettingsDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileMesh(clusterName string, obj *appmesh_k8s_aws_v1beta2.Mesh) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileMeshDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileConfigMap(clusterName string, obj *v1.ConfigMap) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileConfigMapDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileService(clusterName string, obj *v1.Service) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileServiceDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcilePod(clusterName string, obj *v1.Pod) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcilePodDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileNode(clusterName string, obj *v1.Node) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileNodeDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileDeployment(clusterName string, obj *apps_v1.Deployment) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileDeploymentDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileReplicaSet(clusterName string, obj *apps_v1.ReplicaSet) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileReplicaSetDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileDaemonSet(clusterName string, obj *apps_v1.DaemonSet) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileDaemonSetDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterAgentReconcilerImpl) ReconcileStatefulSet(clusterName string, obj *apps_v1.StatefulSet) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterAgentReconcilerImpl) ReconcileStatefulSetDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

// the singleClusterAgentReconciler reconciles events for input resources across clusters
// this private interface is used to ensure that the generated struct implements the intended functions
type singleClusterAgentReconciler interface {
	settings_mesh_gloo_solo_io_v1alpha2_controllers.SettingsReconciler

	appmesh_k8s_aws_v1beta2_controllers.MeshReconciler

	v1_controllers.ConfigMapReconciler
	v1_controllers.ServiceReconciler
	v1_controllers.PodReconciler
	v1_controllers.NodeReconciler

	apps_v1_controllers.DeploymentReconciler
	apps_v1_controllers.ReplicaSetReconciler
	apps_v1_controllers.DaemonSetReconciler
	apps_v1_controllers.StatefulSetReconciler
}

var _ singleClusterAgentReconciler = &singleClusterAgentReconcilerImpl{}

type singleClusterAgentReconcilerImpl struct {
	base input.InputReconciler
}

// register the reconcile func with the manager
// the reconcileInterval, if greater than 0, will limit the number of reconciles
// to one per interval.
func RegisterSingleClusterAgentReconciler(
	ctx context.Context,
	mgr manager.Manager,
	reconcileFunc input.SingleClusterReconcileFunc,
	reconcileInterval time.Duration,
	options reconcile.Options,
	predicates ...predicate.Predicate,
) (input.InputReconciler, error) {

	base := input.NewInputReconciler(
		ctx,
		nil,
		reconcileFunc,
		reconcileInterval,
	)

	r := &singleClusterAgentReconcilerImpl{
		base: base,
	}

	// initialize reconcile loops

	if err := settings_mesh_gloo_solo_io_v1alpha2_controllers.NewSettingsReconcileLoop("Settings", mgr, options).RunSettingsReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}

	if err := appmesh_k8s_aws_v1beta2_controllers.NewMeshReconcileLoop("Mesh", mgr, options).RunMeshReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}

	if err := v1_controllers.NewConfigMapReconcileLoop("ConfigMap", mgr, options).RunConfigMapReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := v1_controllers.NewServiceReconcileLoop("Service", mgr, options).RunServiceReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := v1_controllers.NewPodReconcileLoop("Pod", mgr, options).RunPodReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := v1_controllers.NewNodeReconcileLoop("Node", mgr, options).RunNodeReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}

	if err := apps_v1_controllers.NewDeploymentReconcileLoop("Deployment", mgr, options).RunDeploymentReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := apps_v1_controllers.NewReplicaSetReconcileLoop("ReplicaSet", mgr, options).RunReplicaSetReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := apps_v1_controllers.NewDaemonSetReconcileLoop("DaemonSet", mgr, options).RunDaemonSetReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := apps_v1_controllers.NewStatefulSetReconcileLoop("StatefulSet", mgr, options).RunStatefulSetReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}

	return r.base, nil
}

func (r *singleClusterAgentReconcilerImpl) ReconcileSettings(obj *settings_mesh_gloo_solo_io_v1alpha2.Settings) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileSettingsDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileMesh(obj *appmesh_k8s_aws_v1beta2.Mesh) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileMeshDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileConfigMap(obj *v1.ConfigMap) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileConfigMapDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileService(obj *v1.Service) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileServiceDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcilePod(obj *v1.Pod) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcilePodDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileNode(obj *v1.Node) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileNodeDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileDeployment(obj *apps_v1.Deployment) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileDeploymentDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileReplicaSet(obj *apps_v1.ReplicaSet) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileReplicaSetDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileDaemonSet(obj *apps_v1.DaemonSet) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileDaemonSetDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterAgentReconcilerImpl) ReconcileStatefulSet(obj *apps_v1.StatefulSet) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterAgentReconcilerImpl) ReconcileStatefulSetDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}
