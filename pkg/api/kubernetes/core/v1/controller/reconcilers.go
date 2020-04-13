// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	core_v1 "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconcile Upsert events for the Secret Resource.
// implemented by the user
type SecretReconciler interface {
	ReconcileSecret(obj *core_v1.Secret) (reconcile.Result, error)
}

// Reconcile deletion events for the Secret Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type SecretDeletionReconciler interface {
	ReconcileSecretDeletion(req reconcile.Request)
}

type SecretReconcilerFuncs struct {
	OnReconcileSecret         func(obj *core_v1.Secret) (reconcile.Result, error)
	OnReconcileSecretDeletion func(req reconcile.Request)
}

func (f *SecretReconcilerFuncs) ReconcileSecret(obj *core_v1.Secret) (reconcile.Result, error) {
	if f.OnReconcileSecret == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileSecret(obj)
}

func (f *SecretReconcilerFuncs) ReconcileSecretDeletion(req reconcile.Request) {
	if f.OnReconcileSecretDeletion == nil {
		return
	}
	f.OnReconcileSecretDeletion(req)
}

// Reconcile and finalize the Secret Resource
// implemented by the user
type SecretFinalizer interface {
	SecretReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	SecretFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeSecret(obj *core_v1.Secret) error
}

type SecretReconcileLoop interface {
	RunSecretReconciler(ctx context.Context, rec SecretReconciler, predicates ...predicate.Predicate) error
}

type secretReconcileLoop struct {
	loop reconcile.Loop
}

func NewSecretReconcileLoop(name string, mgr manager.Manager) SecretReconcileLoop {
	return &secretReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Secret{}),
	}
}

func (c *secretReconcileLoop) RunSecretReconciler(ctx context.Context, reconciler SecretReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericSecretReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(SecretFinalizer); ok {
		reconcilerWrapper = genericSecretFinalizer{
			genericSecretReconciler: genericReconciler,
			finalizingReconciler:    finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericSecretHandler implements a generic reconcile.Reconciler
type genericSecretReconciler struct {
	reconciler SecretReconciler
}

func (r genericSecretReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return r.reconciler.ReconcileSecret(obj)
}

func (r genericSecretReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(SecretDeletionReconciler); ok {
		deletionReconciler.ReconcileSecretDeletion(request)
	}
}

// genericSecretFinalizer implements a generic reconcile.FinalizingReconciler
type genericSecretFinalizer struct {
	genericSecretReconciler
	finalizingReconciler SecretFinalizer
}

func (r genericSecretFinalizer) FinalizerName() string {
	return r.finalizingReconciler.SecretFinalizerName()
}

func (r genericSecretFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeSecret(obj)
}

// Reconcile Upsert events for the Service Resource.
// implemented by the user
type ServiceReconciler interface {
	ReconcileService(obj *core_v1.Service) (reconcile.Result, error)
}

// Reconcile deletion events for the Service Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type ServiceDeletionReconciler interface {
	ReconcileServiceDeletion(req reconcile.Request)
}

type ServiceReconcilerFuncs struct {
	OnReconcileService         func(obj *core_v1.Service) (reconcile.Result, error)
	OnReconcileServiceDeletion func(req reconcile.Request)
}

func (f *ServiceReconcilerFuncs) ReconcileService(obj *core_v1.Service) (reconcile.Result, error) {
	if f.OnReconcileService == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileService(obj)
}

func (f *ServiceReconcilerFuncs) ReconcileServiceDeletion(req reconcile.Request) {
	if f.OnReconcileServiceDeletion == nil {
		return
	}
	f.OnReconcileServiceDeletion(req)
}

// Reconcile and finalize the Service Resource
// implemented by the user
type ServiceFinalizer interface {
	ServiceReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	ServiceFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeService(obj *core_v1.Service) error
}

type ServiceReconcileLoop interface {
	RunServiceReconciler(ctx context.Context, rec ServiceReconciler, predicates ...predicate.Predicate) error
}

type serviceReconcileLoop struct {
	loop reconcile.Loop
}

func NewServiceReconcileLoop(name string, mgr manager.Manager) ServiceReconcileLoop {
	return &serviceReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Service{}),
	}
}

func (c *serviceReconcileLoop) RunServiceReconciler(ctx context.Context, reconciler ServiceReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericServiceReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(ServiceFinalizer); ok {
		reconcilerWrapper = genericServiceFinalizer{
			genericServiceReconciler: genericReconciler,
			finalizingReconciler:     finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericServiceHandler implements a generic reconcile.Reconciler
type genericServiceReconciler struct {
	reconciler ServiceReconciler
}

func (r genericServiceReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return r.reconciler.ReconcileService(obj)
}

func (r genericServiceReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(ServiceDeletionReconciler); ok {
		deletionReconciler.ReconcileServiceDeletion(request)
	}
}

// genericServiceFinalizer implements a generic reconcile.FinalizingReconciler
type genericServiceFinalizer struct {
	genericServiceReconciler
	finalizingReconciler ServiceFinalizer
}

func (r genericServiceFinalizer) FinalizerName() string {
	return r.finalizingReconciler.ServiceFinalizerName()
}

func (r genericServiceFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeService(obj)
}

// Reconcile Upsert events for the Pod Resource.
// implemented by the user
type PodReconciler interface {
	ReconcilePod(obj *core_v1.Pod) (reconcile.Result, error)
}

// Reconcile deletion events for the Pod Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type PodDeletionReconciler interface {
	ReconcilePodDeletion(req reconcile.Request)
}

type PodReconcilerFuncs struct {
	OnReconcilePod         func(obj *core_v1.Pod) (reconcile.Result, error)
	OnReconcilePodDeletion func(req reconcile.Request)
}

func (f *PodReconcilerFuncs) ReconcilePod(obj *core_v1.Pod) (reconcile.Result, error) {
	if f.OnReconcilePod == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcilePod(obj)
}

func (f *PodReconcilerFuncs) ReconcilePodDeletion(req reconcile.Request) {
	if f.OnReconcilePodDeletion == nil {
		return
	}
	f.OnReconcilePodDeletion(req)
}

// Reconcile and finalize the Pod Resource
// implemented by the user
type PodFinalizer interface {
	PodReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	PodFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizePod(obj *core_v1.Pod) error
}

type PodReconcileLoop interface {
	RunPodReconciler(ctx context.Context, rec PodReconciler, predicates ...predicate.Predicate) error
}

type podReconcileLoop struct {
	loop reconcile.Loop
}

func NewPodReconcileLoop(name string, mgr manager.Manager) PodReconcileLoop {
	return &podReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Pod{}),
	}
}

func (c *podReconcileLoop) RunPodReconciler(ctx context.Context, reconciler PodReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericPodReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(PodFinalizer); ok {
		reconcilerWrapper = genericPodFinalizer{
			genericPodReconciler: genericReconciler,
			finalizingReconciler: finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericPodHandler implements a generic reconcile.Reconciler
type genericPodReconciler struct {
	reconciler PodReconciler
}

func (r genericPodReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return r.reconciler.ReconcilePod(obj)
}

func (r genericPodReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(PodDeletionReconciler); ok {
		deletionReconciler.ReconcilePodDeletion(request)
	}
}

// genericPodFinalizer implements a generic reconcile.FinalizingReconciler
type genericPodFinalizer struct {
	genericPodReconciler
	finalizingReconciler PodFinalizer
}

func (r genericPodFinalizer) FinalizerName() string {
	return r.finalizingReconciler.PodFinalizerName()
}

func (r genericPodFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizePod(obj)
}
