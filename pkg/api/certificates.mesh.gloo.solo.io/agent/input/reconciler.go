// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./reconciler.go -destination mocks/reconciler.go

// The Input Reconciler calls a simple func() error whenever a
// storage event is received for any of:
// * CertificatesMeshGlooSoloIov1Alpha2IssuedCertificates
// * CertificatesMeshGlooSoloIov1Alpha2CertificateRequests
// * CertificatesMeshGlooSoloIov1Alpha2PodBounceDirectives
// * V1Secrets
// * V1Pods
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

	certificates_mesh_gloo_solo_io_v1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2"
	certificates_mesh_gloo_solo_io_v1alpha2_controllers "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2/controller"

	v1_controllers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	v1 "k8s.io/api/core/v1"
)

// the multiClusterReconciler reconciles events for input resources across clusters
// this private interface is used to ensure that the generated struct implements the intended functions
type multiClusterReconciler interface {
	certificates_mesh_gloo_solo_io_v1alpha2_controllers.MulticlusterIssuedCertificateReconciler
	certificates_mesh_gloo_solo_io_v1alpha2_controllers.MulticlusterCertificateRequestReconciler
	certificates_mesh_gloo_solo_io_v1alpha2_controllers.MulticlusterPodBounceDirectiveReconciler

	v1_controllers.MulticlusterSecretReconciler
	v1_controllers.MulticlusterPodReconciler
}

var _ multiClusterReconciler = &multiClusterReconcilerImpl{}

type multiClusterReconcilerImpl struct {
	base input.InputReconciler
}

// Options for reconciling a snapshot
type ReconcileOptions struct {

	// Options for reconciling CertificatesMeshGlooSoloIov1Alpha2IssuedCertificates
	CertificatesMeshGlooSoloIov1Alpha2IssuedCertificates reconcile.Options
	// Options for reconciling CertificatesMeshGlooSoloIov1Alpha2CertificateRequests
	CertificatesMeshGlooSoloIov1Alpha2CertificateRequests reconcile.Options
	// Options for reconciling CertificatesMeshGlooSoloIov1Alpha2PodBounceDirectives
	CertificatesMeshGlooSoloIov1Alpha2PodBounceDirectives reconcile.Options

	// Options for reconciling V1Secrets
	V1Secrets reconcile.Options
	// Options for reconciling V1Pods
	V1Pods reconcile.Options
}

// register the reconcile func with the cluster watcher
// the reconcileInterval, if greater than 0, will limit the number of reconciles
// to one per interval.
func RegisterMultiClusterReconciler(
	ctx context.Context,
	clusters multicluster.ClusterWatcher,
	reconcileFunc input.MultiClusterReconcileFunc,
	reconcileInterval time.Duration,
	options ReconcileOptions,
	predicates ...predicate.Predicate,
) input.InputReconciler {

	base := input.NewInputReconciler(
		ctx,
		reconcileFunc,
		nil,
		reconcileInterval,
	)

	r := &multiClusterReconcilerImpl{
		base: base,
	}

	// initialize reconcile loops

	certificates_mesh_gloo_solo_io_v1alpha2_controllers.NewMulticlusterIssuedCertificateReconcileLoop("IssuedCertificate", clusters, options.CertificatesMeshGlooSoloIov1Alpha2IssuedCertificates).AddMulticlusterIssuedCertificateReconciler(ctx, r, predicates...)

	certificates_mesh_gloo_solo_io_v1alpha2_controllers.NewMulticlusterCertificateRequestReconcileLoop("CertificateRequest", clusters, options.CertificatesMeshGlooSoloIov1Alpha2CertificateRequests).AddMulticlusterCertificateRequestReconciler(ctx, r, predicates...)

	certificates_mesh_gloo_solo_io_v1alpha2_controllers.NewMulticlusterPodBounceDirectiveReconcileLoop("PodBounceDirective", clusters, options.CertificatesMeshGlooSoloIov1Alpha2PodBounceDirectives).AddMulticlusterPodBounceDirectiveReconciler(ctx, r, predicates...)

	v1_controllers.NewMulticlusterSecretReconcileLoop("Secret", clusters, options.V1Secrets).AddMulticlusterSecretReconciler(ctx, r, predicates...)

	v1_controllers.NewMulticlusterPodReconcileLoop("Pod", clusters, options.V1Pods).AddMulticlusterPodReconciler(ctx, r, predicates...)
	return r.base
}

func (r *multiClusterReconcilerImpl) ReconcileIssuedCertificate(clusterName string, obj *certificates_mesh_gloo_solo_io_v1alpha2.IssuedCertificate) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterReconcilerImpl) ReconcileIssuedCertificateDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterReconcilerImpl) ReconcileCertificateRequest(clusterName string, obj *certificates_mesh_gloo_solo_io_v1alpha2.CertificateRequest) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterReconcilerImpl) ReconcileCertificateRequestDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterReconcilerImpl) ReconcilePodBounceDirective(clusterName string, obj *certificates_mesh_gloo_solo_io_v1alpha2.PodBounceDirective) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterReconcilerImpl) ReconcilePodBounceDirectiveDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterReconcilerImpl) ReconcileSecret(clusterName string, obj *v1.Secret) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterReconcilerImpl) ReconcileSecretDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

func (r *multiClusterReconcilerImpl) ReconcilePod(clusterName string, obj *v1.Pod) (reconcile.Result, error) {
	obj.ClusterName = clusterName
	return r.base.ReconcileRemoteGeneric(obj)
}

func (r *multiClusterReconcilerImpl) ReconcilePodDeletion(clusterName string, obj reconcile.Request) error {
	ref := &sk_core_v1.ClusterObjectRef{
		Name:        obj.Name,
		Namespace:   obj.Namespace,
		ClusterName: clusterName,
	}
	_, err := r.base.ReconcileRemoteGeneric(ref)
	return err
}

// the singleClusterReconciler reconciles events for input resources across clusters
// this private interface is used to ensure that the generated struct implements the intended functions
type singleClusterReconciler interface {
	certificates_mesh_gloo_solo_io_v1alpha2_controllers.IssuedCertificateReconciler
	certificates_mesh_gloo_solo_io_v1alpha2_controllers.CertificateRequestReconciler
	certificates_mesh_gloo_solo_io_v1alpha2_controllers.PodBounceDirectiveReconciler

	v1_controllers.SecretReconciler
	v1_controllers.PodReconciler
}

var _ singleClusterReconciler = &singleClusterReconcilerImpl{}

type singleClusterReconcilerImpl struct {
	base input.InputReconciler
}

// register the reconcile func with the manager
// the reconcileInterval, if greater than 0, will limit the number of reconciles
// to one per interval.
func RegisterSingleClusterReconciler(
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

	r := &singleClusterReconcilerImpl{
		base: base,
	}

	// initialize reconcile loops

	if err := certificates_mesh_gloo_solo_io_v1alpha2_controllers.NewIssuedCertificateReconcileLoop("IssuedCertificate", mgr, options).RunIssuedCertificateReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := certificates_mesh_gloo_solo_io_v1alpha2_controllers.NewCertificateRequestReconcileLoop("CertificateRequest", mgr, options).RunCertificateRequestReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := certificates_mesh_gloo_solo_io_v1alpha2_controllers.NewPodBounceDirectiveReconcileLoop("PodBounceDirective", mgr, options).RunPodBounceDirectiveReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}

	if err := v1_controllers.NewSecretReconcileLoop("Secret", mgr, options).RunSecretReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}
	if err := v1_controllers.NewPodReconcileLoop("Pod", mgr, options).RunPodReconciler(ctx, r, predicates...); err != nil {
		return nil, err
	}

	return r.base, nil
}

func (r *singleClusterReconcilerImpl) ReconcileIssuedCertificate(obj *certificates_mesh_gloo_solo_io_v1alpha2.IssuedCertificate) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterReconcilerImpl) ReconcileIssuedCertificateDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterReconcilerImpl) ReconcileCertificateRequest(obj *certificates_mesh_gloo_solo_io_v1alpha2.CertificateRequest) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterReconcilerImpl) ReconcileCertificateRequestDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterReconcilerImpl) ReconcilePodBounceDirective(obj *certificates_mesh_gloo_solo_io_v1alpha2.PodBounceDirective) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterReconcilerImpl) ReconcilePodBounceDirectiveDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterReconcilerImpl) ReconcileSecret(obj *v1.Secret) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterReconcilerImpl) ReconcileSecretDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}

func (r *singleClusterReconcilerImpl) ReconcilePod(obj *v1.Pod) (reconcile.Result, error) {
	return r.base.ReconcileLocalGeneric(obj)
}

func (r *singleClusterReconcilerImpl) ReconcilePodDeletion(obj reconcile.Request) error {
	ref := &sk_core_v1.ObjectRef{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
	_, err := r.base.ReconcileLocalGeneric(ref)
	return err
}
