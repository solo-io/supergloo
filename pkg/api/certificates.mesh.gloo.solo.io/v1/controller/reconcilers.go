// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./reconcilers.go -destination mocks/reconcilers.go

// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	certificates_mesh_gloo_solo_io_v1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconcile Upsert events for the IssuedCertificate Resource.
// implemented by the user
type IssuedCertificateReconciler interface {
	ReconcileIssuedCertificate(obj *certificates_mesh_gloo_solo_io_v1.IssuedCertificate) (reconcile.Result, error)
}

// Reconcile deletion events for the IssuedCertificate Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type IssuedCertificateDeletionReconciler interface {
	ReconcileIssuedCertificateDeletion(req reconcile.Request) error
}

type IssuedCertificateReconcilerFuncs struct {
	OnReconcileIssuedCertificate         func(obj *certificates_mesh_gloo_solo_io_v1.IssuedCertificate) (reconcile.Result, error)
	OnReconcileIssuedCertificateDeletion func(req reconcile.Request) error
}

func (f *IssuedCertificateReconcilerFuncs) ReconcileIssuedCertificate(obj *certificates_mesh_gloo_solo_io_v1.IssuedCertificate) (reconcile.Result, error) {
	if f.OnReconcileIssuedCertificate == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileIssuedCertificate(obj)
}

func (f *IssuedCertificateReconcilerFuncs) ReconcileIssuedCertificateDeletion(req reconcile.Request) error {
	if f.OnReconcileIssuedCertificateDeletion == nil {
		return nil
	}
	return f.OnReconcileIssuedCertificateDeletion(req)
}

// Reconcile and finalize the IssuedCertificate Resource
// implemented by the user
type IssuedCertificateFinalizer interface {
	IssuedCertificateReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	IssuedCertificateFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeIssuedCertificate(obj *certificates_mesh_gloo_solo_io_v1.IssuedCertificate) error
}

type IssuedCertificateReconcileLoop interface {
	RunIssuedCertificateReconciler(ctx context.Context, rec IssuedCertificateReconciler, predicates ...predicate.Predicate) error
}

type issuedCertificateReconcileLoop struct {
	loop reconcile.Loop
}

func NewIssuedCertificateReconcileLoop(name string, mgr manager.Manager, options reconcile.Options) IssuedCertificateReconcileLoop {
	return &issuedCertificateReconcileLoop{
		// empty cluster indicates this reconciler is built for the local cluster
		loop: reconcile.NewLoop(name, "", mgr, &certificates_mesh_gloo_solo_io_v1.IssuedCertificate{}, options),
	}
}

func (c *issuedCertificateReconcileLoop) RunIssuedCertificateReconciler(ctx context.Context, reconciler IssuedCertificateReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericIssuedCertificateReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(IssuedCertificateFinalizer); ok {
		reconcilerWrapper = genericIssuedCertificateFinalizer{
			genericIssuedCertificateReconciler: genericReconciler,
			finalizingReconciler:               finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericIssuedCertificateHandler implements a generic reconcile.Reconciler
type genericIssuedCertificateReconciler struct {
	reconciler IssuedCertificateReconciler
}

func (r genericIssuedCertificateReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.IssuedCertificate)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: IssuedCertificate handler received event for %T", object)
	}
	return r.reconciler.ReconcileIssuedCertificate(obj)
}

func (r genericIssuedCertificateReconciler) ReconcileDeletion(request reconcile.Request) error {
	if deletionReconciler, ok := r.reconciler.(IssuedCertificateDeletionReconciler); ok {
		return deletionReconciler.ReconcileIssuedCertificateDeletion(request)
	}
	return nil
}

// genericIssuedCertificateFinalizer implements a generic reconcile.FinalizingReconciler
type genericIssuedCertificateFinalizer struct {
	genericIssuedCertificateReconciler
	finalizingReconciler IssuedCertificateFinalizer
}

func (r genericIssuedCertificateFinalizer) FinalizerName() string {
	return r.finalizingReconciler.IssuedCertificateFinalizerName()
}

func (r genericIssuedCertificateFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.IssuedCertificate)
	if !ok {
		return errors.Errorf("internal error: IssuedCertificate handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeIssuedCertificate(obj)
}

// Reconcile Upsert events for the CertificateRequest Resource.
// implemented by the user
type CertificateRequestReconciler interface {
	ReconcileCertificateRequest(obj *certificates_mesh_gloo_solo_io_v1.CertificateRequest) (reconcile.Result, error)
}

// Reconcile deletion events for the CertificateRequest Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type CertificateRequestDeletionReconciler interface {
	ReconcileCertificateRequestDeletion(req reconcile.Request) error
}

type CertificateRequestReconcilerFuncs struct {
	OnReconcileCertificateRequest         func(obj *certificates_mesh_gloo_solo_io_v1.CertificateRequest) (reconcile.Result, error)
	OnReconcileCertificateRequestDeletion func(req reconcile.Request) error
}

func (f *CertificateRequestReconcilerFuncs) ReconcileCertificateRequest(obj *certificates_mesh_gloo_solo_io_v1.CertificateRequest) (reconcile.Result, error) {
	if f.OnReconcileCertificateRequest == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileCertificateRequest(obj)
}

func (f *CertificateRequestReconcilerFuncs) ReconcileCertificateRequestDeletion(req reconcile.Request) error {
	if f.OnReconcileCertificateRequestDeletion == nil {
		return nil
	}
	return f.OnReconcileCertificateRequestDeletion(req)
}

// Reconcile and finalize the CertificateRequest Resource
// implemented by the user
type CertificateRequestFinalizer interface {
	CertificateRequestReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	CertificateRequestFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeCertificateRequest(obj *certificates_mesh_gloo_solo_io_v1.CertificateRequest) error
}

type CertificateRequestReconcileLoop interface {
	RunCertificateRequestReconciler(ctx context.Context, rec CertificateRequestReconciler, predicates ...predicate.Predicate) error
}

type certificateRequestReconcileLoop struct {
	loop reconcile.Loop
}

func NewCertificateRequestReconcileLoop(name string, mgr manager.Manager, options reconcile.Options) CertificateRequestReconcileLoop {
	return &certificateRequestReconcileLoop{
		// empty cluster indicates this reconciler is built for the local cluster
		loop: reconcile.NewLoop(name, "", mgr, &certificates_mesh_gloo_solo_io_v1.CertificateRequest{}, options),
	}
}

func (c *certificateRequestReconcileLoop) RunCertificateRequestReconciler(ctx context.Context, reconciler CertificateRequestReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericCertificateRequestReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(CertificateRequestFinalizer); ok {
		reconcilerWrapper = genericCertificateRequestFinalizer{
			genericCertificateRequestReconciler: genericReconciler,
			finalizingReconciler:                finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericCertificateRequestHandler implements a generic reconcile.Reconciler
type genericCertificateRequestReconciler struct {
	reconciler CertificateRequestReconciler
}

func (r genericCertificateRequestReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.CertificateRequest)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: CertificateRequest handler received event for %T", object)
	}
	return r.reconciler.ReconcileCertificateRequest(obj)
}

func (r genericCertificateRequestReconciler) ReconcileDeletion(request reconcile.Request) error {
	if deletionReconciler, ok := r.reconciler.(CertificateRequestDeletionReconciler); ok {
		return deletionReconciler.ReconcileCertificateRequestDeletion(request)
	}
	return nil
}

// genericCertificateRequestFinalizer implements a generic reconcile.FinalizingReconciler
type genericCertificateRequestFinalizer struct {
	genericCertificateRequestReconciler
	finalizingReconciler CertificateRequestFinalizer
}

func (r genericCertificateRequestFinalizer) FinalizerName() string {
	return r.finalizingReconciler.CertificateRequestFinalizerName()
}

func (r genericCertificateRequestFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.CertificateRequest)
	if !ok {
		return errors.Errorf("internal error: CertificateRequest handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeCertificateRequest(obj)
}

// Reconcile Upsert events for the PodBounceDirective Resource.
// implemented by the user
type PodBounceDirectiveReconciler interface {
	ReconcilePodBounceDirective(obj *certificates_mesh_gloo_solo_io_v1.PodBounceDirective) (reconcile.Result, error)
}

// Reconcile deletion events for the PodBounceDirective Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type PodBounceDirectiveDeletionReconciler interface {
	ReconcilePodBounceDirectiveDeletion(req reconcile.Request) error
}

type PodBounceDirectiveReconcilerFuncs struct {
	OnReconcilePodBounceDirective         func(obj *certificates_mesh_gloo_solo_io_v1.PodBounceDirective) (reconcile.Result, error)
	OnReconcilePodBounceDirectiveDeletion func(req reconcile.Request) error
}

func (f *PodBounceDirectiveReconcilerFuncs) ReconcilePodBounceDirective(obj *certificates_mesh_gloo_solo_io_v1.PodBounceDirective) (reconcile.Result, error) {
	if f.OnReconcilePodBounceDirective == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcilePodBounceDirective(obj)
}

func (f *PodBounceDirectiveReconcilerFuncs) ReconcilePodBounceDirectiveDeletion(req reconcile.Request) error {
	if f.OnReconcilePodBounceDirectiveDeletion == nil {
		return nil
	}
	return f.OnReconcilePodBounceDirectiveDeletion(req)
}

// Reconcile and finalize the PodBounceDirective Resource
// implemented by the user
type PodBounceDirectiveFinalizer interface {
	PodBounceDirectiveReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	PodBounceDirectiveFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizePodBounceDirective(obj *certificates_mesh_gloo_solo_io_v1.PodBounceDirective) error
}

type PodBounceDirectiveReconcileLoop interface {
	RunPodBounceDirectiveReconciler(ctx context.Context, rec PodBounceDirectiveReconciler, predicates ...predicate.Predicate) error
}

type podBounceDirectiveReconcileLoop struct {
	loop reconcile.Loop
}

func NewPodBounceDirectiveReconcileLoop(name string, mgr manager.Manager, options reconcile.Options) PodBounceDirectiveReconcileLoop {
	return &podBounceDirectiveReconcileLoop{
		// empty cluster indicates this reconciler is built for the local cluster
		loop: reconcile.NewLoop(name, "", mgr, &certificates_mesh_gloo_solo_io_v1.PodBounceDirective{}, options),
	}
}

func (c *podBounceDirectiveReconcileLoop) RunPodBounceDirectiveReconciler(ctx context.Context, reconciler PodBounceDirectiveReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericPodBounceDirectiveReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(PodBounceDirectiveFinalizer); ok {
		reconcilerWrapper = genericPodBounceDirectiveFinalizer{
			genericPodBounceDirectiveReconciler: genericReconciler,
			finalizingReconciler:                finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericPodBounceDirectiveHandler implements a generic reconcile.Reconciler
type genericPodBounceDirectiveReconciler struct {
	reconciler PodBounceDirectiveReconciler
}

func (r genericPodBounceDirectiveReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.PodBounceDirective)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: PodBounceDirective handler received event for %T", object)
	}
	return r.reconciler.ReconcilePodBounceDirective(obj)
}

func (r genericPodBounceDirectiveReconciler) ReconcileDeletion(request reconcile.Request) error {
	if deletionReconciler, ok := r.reconciler.(PodBounceDirectiveDeletionReconciler); ok {
		return deletionReconciler.ReconcilePodBounceDirectiveDeletion(request)
	}
	return nil
}

// genericPodBounceDirectiveFinalizer implements a generic reconcile.FinalizingReconciler
type genericPodBounceDirectiveFinalizer struct {
	genericPodBounceDirectiveReconciler
	finalizingReconciler PodBounceDirectiveFinalizer
}

func (r genericPodBounceDirectiveFinalizer) FinalizerName() string {
	return r.finalizingReconciler.PodBounceDirectiveFinalizerName()
}

func (r genericPodBounceDirectiveFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.PodBounceDirective)
	if !ok {
		return errors.Errorf("internal error: PodBounceDirective handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizePodBounceDirective(obj)
}

// Reconcile Upsert events for the CertificateRotation Resource.
// implemented by the user
type CertificateRotationReconciler interface {
	ReconcileCertificateRotation(obj *certificates_mesh_gloo_solo_io_v1.CertificateRotation) (reconcile.Result, error)
}

// Reconcile deletion events for the CertificateRotation Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type CertificateRotationDeletionReconciler interface {
	ReconcileCertificateRotationDeletion(req reconcile.Request) error
}

type CertificateRotationReconcilerFuncs struct {
	OnReconcileCertificateRotation         func(obj *certificates_mesh_gloo_solo_io_v1.CertificateRotation) (reconcile.Result, error)
	OnReconcileCertificateRotationDeletion func(req reconcile.Request) error
}

func (f *CertificateRotationReconcilerFuncs) ReconcileCertificateRotation(obj *certificates_mesh_gloo_solo_io_v1.CertificateRotation) (reconcile.Result, error) {
	if f.OnReconcileCertificateRotation == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileCertificateRotation(obj)
}

func (f *CertificateRotationReconcilerFuncs) ReconcileCertificateRotationDeletion(req reconcile.Request) error {
	if f.OnReconcileCertificateRotationDeletion == nil {
		return nil
	}
	return f.OnReconcileCertificateRotationDeletion(req)
}

// Reconcile and finalize the CertificateRotation Resource
// implemented by the user
type CertificateRotationFinalizer interface {
	CertificateRotationReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	CertificateRotationFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeCertificateRotation(obj *certificates_mesh_gloo_solo_io_v1.CertificateRotation) error
}

type CertificateRotationReconcileLoop interface {
	RunCertificateRotationReconciler(ctx context.Context, rec CertificateRotationReconciler, predicates ...predicate.Predicate) error
}

type certificateRotationReconcileLoop struct {
	loop reconcile.Loop
}

func NewCertificateRotationReconcileLoop(name string, mgr manager.Manager, options reconcile.Options) CertificateRotationReconcileLoop {
	return &certificateRotationReconcileLoop{
		// empty cluster indicates this reconciler is built for the local cluster
		loop: reconcile.NewLoop(name, "", mgr, &certificates_mesh_gloo_solo_io_v1.CertificateRotation{}, options),
	}
}

func (c *certificateRotationReconcileLoop) RunCertificateRotationReconciler(ctx context.Context, reconciler CertificateRotationReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericCertificateRotationReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(CertificateRotationFinalizer); ok {
		reconcilerWrapper = genericCertificateRotationFinalizer{
			genericCertificateRotationReconciler: genericReconciler,
			finalizingReconciler:                 finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericCertificateRotationHandler implements a generic reconcile.Reconciler
type genericCertificateRotationReconciler struct {
	reconciler CertificateRotationReconciler
}

func (r genericCertificateRotationReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.CertificateRotation)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: CertificateRotation handler received event for %T", object)
	}
	return r.reconciler.ReconcileCertificateRotation(obj)
}

func (r genericCertificateRotationReconciler) ReconcileDeletion(request reconcile.Request) error {
	if deletionReconciler, ok := r.reconciler.(CertificateRotationDeletionReconciler); ok {
		return deletionReconciler.ReconcileCertificateRotationDeletion(request)
	}
	return nil
}

// genericCertificateRotationFinalizer implements a generic reconcile.FinalizingReconciler
type genericCertificateRotationFinalizer struct {
	genericCertificateRotationReconciler
	finalizingReconciler CertificateRotationFinalizer
}

func (r genericCertificateRotationFinalizer) FinalizerName() string {
	return r.finalizingReconciler.CertificateRotationFinalizerName()
}

func (r genericCertificateRotationFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*certificates_mesh_gloo_solo_io_v1.CertificateRotation)
	if !ok {
		return errors.Errorf("internal error: CertificateRotation handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeCertificateRotation(obj)
}
