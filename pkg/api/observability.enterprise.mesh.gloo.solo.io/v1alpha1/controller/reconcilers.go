// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./reconcilers.go -destination mocks/reconcilers.go

// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	observability_enterprise_mesh_gloo_solo_io_v1alpha1 "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconcile Upsert events for the AccessLogCollection Resource.
// implemented by the user
type AccessLogCollectionReconciler interface {
	ReconcileAccessLogCollection(obj *observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogCollection) (reconcile.Result, error)
}

// Reconcile deletion events for the AccessLogCollection Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type AccessLogCollectionDeletionReconciler interface {
	ReconcileAccessLogCollectionDeletion(req reconcile.Request) error
}

type AccessLogCollectionReconcilerFuncs struct {
	OnReconcileAccessLogCollection         func(obj *observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogCollection) (reconcile.Result, error)
	OnReconcileAccessLogCollectionDeletion func(req reconcile.Request) error
}

func (f *AccessLogCollectionReconcilerFuncs) ReconcileAccessLogCollection(obj *observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogCollection) (reconcile.Result, error) {
	if f.OnReconcileAccessLogCollection == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileAccessLogCollection(obj)
}

func (f *AccessLogCollectionReconcilerFuncs) ReconcileAccessLogCollectionDeletion(req reconcile.Request) error {
	if f.OnReconcileAccessLogCollectionDeletion == nil {
		return nil
	}
	return f.OnReconcileAccessLogCollectionDeletion(req)
}

// Reconcile and finalize the AccessLogCollection Resource
// implemented by the user
type AccessLogCollectionFinalizer interface {
	AccessLogCollectionReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	AccessLogCollectionFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeAccessLogCollection(obj *observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogCollection) error
}

type AccessLogCollectionReconcileLoop interface {
	RunAccessLogCollectionReconciler(ctx context.Context, rec AccessLogCollectionReconciler, predicates ...predicate.Predicate) error
}

type accessLogCollectionReconcileLoop struct {
	loop reconcile.Loop
}

func NewAccessLogCollectionReconcileLoop(name string, mgr manager.Manager, options reconcile.Options) AccessLogCollectionReconcileLoop {
	return &accessLogCollectionReconcileLoop{
		// empty cluster indicates this reconciler is built for the local cluster
		loop: reconcile.NewLoop(name, "", mgr, &observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogCollection{}, options),
	}
}

func (c *accessLogCollectionReconcileLoop) RunAccessLogCollectionReconciler(ctx context.Context, reconciler AccessLogCollectionReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericAccessLogCollectionReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(AccessLogCollectionFinalizer); ok {
		reconcilerWrapper = genericAccessLogCollectionFinalizer{
			genericAccessLogCollectionReconciler: genericReconciler,
			finalizingReconciler:                 finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericAccessLogCollectionHandler implements a generic reconcile.Reconciler
type genericAccessLogCollectionReconciler struct {
	reconciler AccessLogCollectionReconciler
}

func (r genericAccessLogCollectionReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogCollection)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: AccessLogCollection handler received event for %T", object)
	}
	return r.reconciler.ReconcileAccessLogCollection(obj)
}

func (r genericAccessLogCollectionReconciler) ReconcileDeletion(request reconcile.Request) error {
	if deletionReconciler, ok := r.reconciler.(AccessLogCollectionDeletionReconciler); ok {
		return deletionReconciler.ReconcileAccessLogCollectionDeletion(request)
	}
	return nil
}

// genericAccessLogCollectionFinalizer implements a generic reconcile.FinalizingReconciler
type genericAccessLogCollectionFinalizer struct {
	genericAccessLogCollectionReconciler
	finalizingReconciler AccessLogCollectionFinalizer
}

func (r genericAccessLogCollectionFinalizer) FinalizerName() string {
	return r.finalizingReconciler.AccessLogCollectionFinalizerName()
}

func (r genericAccessLogCollectionFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*observability_enterprise_mesh_gloo_solo_io_v1alpha1.AccessLogCollection)
	if !ok {
		return errors.Errorf("internal error: AccessLogCollection handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeAccessLogCollection(obj)
}
