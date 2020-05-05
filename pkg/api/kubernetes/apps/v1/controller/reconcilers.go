// Code generated by skv2. DO NOT EDIT.

// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	apps_v1 "k8s.io/api/apps/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconcile Upsert events for the Deployment Resource.
// implemented by the user
type DeploymentReconciler interface {
	ReconcileDeployment(obj *apps_v1.Deployment) (reconcile.Result, error)
}

// Reconcile deletion events for the Deployment Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type DeploymentDeletionReconciler interface {
	ReconcileDeploymentDeletion(req reconcile.Request)
}

type DeploymentReconcilerFuncs struct {
	OnReconcileDeployment         func(obj *apps_v1.Deployment) (reconcile.Result, error)
	OnReconcileDeploymentDeletion func(req reconcile.Request)
}

func (f *DeploymentReconcilerFuncs) ReconcileDeployment(obj *apps_v1.Deployment) (reconcile.Result, error) {
	if f.OnReconcileDeployment == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileDeployment(obj)
}

func (f *DeploymentReconcilerFuncs) ReconcileDeploymentDeletion(req reconcile.Request) {
	if f.OnReconcileDeploymentDeletion == nil {
		return
	}
	f.OnReconcileDeploymentDeletion(req)
}

// Reconcile and finalize the Deployment Resource
// implemented by the user
type DeploymentFinalizer interface {
	DeploymentReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	DeploymentFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeDeployment(obj *apps_v1.Deployment) error
}

type DeploymentReconcileLoop interface {
	RunDeploymentReconciler(ctx context.Context, rec DeploymentReconciler, predicates ...predicate.Predicate) error
}

type deploymentReconcileLoop struct {
	loop reconcile.Loop
}

func NewDeploymentReconcileLoop(name string, mgr manager.Manager) DeploymentReconcileLoop {
	return &deploymentReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &apps_v1.Deployment{}),
	}
}

func (c *deploymentReconcileLoop) RunDeploymentReconciler(ctx context.Context, reconciler DeploymentReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericDeploymentReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(DeploymentFinalizer); ok {
		reconcilerWrapper = genericDeploymentFinalizer{
			genericDeploymentReconciler: genericReconciler,
			finalizingReconciler:        finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericDeploymentHandler implements a generic reconcile.Reconciler
type genericDeploymentReconciler struct {
	reconciler DeploymentReconciler
}

func (r genericDeploymentReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*apps_v1.Deployment)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Deployment handler received event for %T", object)
	}
	return r.reconciler.ReconcileDeployment(obj)
}

func (r genericDeploymentReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(DeploymentDeletionReconciler); ok {
		deletionReconciler.ReconcileDeploymentDeletion(request)
	}
}

// genericDeploymentFinalizer implements a generic reconcile.FinalizingReconciler
type genericDeploymentFinalizer struct {
	genericDeploymentReconciler
	finalizingReconciler DeploymentFinalizer
}

func (r genericDeploymentFinalizer) FinalizerName() string {
	return r.finalizingReconciler.DeploymentFinalizerName()
}

func (r genericDeploymentFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeDeployment(obj)
}

// Reconcile Upsert events for the ReplicaSet Resource.
// implemented by the user
type ReplicaSetReconciler interface {
	ReconcileReplicaSet(obj *apps_v1.ReplicaSet) (reconcile.Result, error)
}

// Reconcile deletion events for the ReplicaSet Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type ReplicaSetDeletionReconciler interface {
	ReconcileReplicaSetDeletion(req reconcile.Request)
}

type ReplicaSetReconcilerFuncs struct {
	OnReconcileReplicaSet         func(obj *apps_v1.ReplicaSet) (reconcile.Result, error)
	OnReconcileReplicaSetDeletion func(req reconcile.Request)
}

func (f *ReplicaSetReconcilerFuncs) ReconcileReplicaSet(obj *apps_v1.ReplicaSet) (reconcile.Result, error) {
	if f.OnReconcileReplicaSet == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileReplicaSet(obj)
}

func (f *ReplicaSetReconcilerFuncs) ReconcileReplicaSetDeletion(req reconcile.Request) {
	if f.OnReconcileReplicaSetDeletion == nil {
		return
	}
	f.OnReconcileReplicaSetDeletion(req)
}

// Reconcile and finalize the ReplicaSet Resource
// implemented by the user
type ReplicaSetFinalizer interface {
	ReplicaSetReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	ReplicaSetFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeReplicaSet(obj *apps_v1.ReplicaSet) error
}

type ReplicaSetReconcileLoop interface {
	RunReplicaSetReconciler(ctx context.Context, rec ReplicaSetReconciler, predicates ...predicate.Predicate) error
}

type replicaSetReconcileLoop struct {
	loop reconcile.Loop
}

func NewReplicaSetReconcileLoop(name string, mgr manager.Manager) ReplicaSetReconcileLoop {
	return &replicaSetReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &apps_v1.ReplicaSet{}),
	}
}

func (c *replicaSetReconcileLoop) RunReplicaSetReconciler(ctx context.Context, reconciler ReplicaSetReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericReplicaSetReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(ReplicaSetFinalizer); ok {
		reconcilerWrapper = genericReplicaSetFinalizer{
			genericReplicaSetReconciler: genericReconciler,
			finalizingReconciler:        finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericReplicaSetHandler implements a generic reconcile.Reconciler
type genericReplicaSetReconciler struct {
	reconciler ReplicaSetReconciler
}

func (r genericReplicaSetReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*apps_v1.ReplicaSet)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: ReplicaSet handler received event for %T", object)
	}
	return r.reconciler.ReconcileReplicaSet(obj)
}

func (r genericReplicaSetReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(ReplicaSetDeletionReconciler); ok {
		deletionReconciler.ReconcileReplicaSetDeletion(request)
	}
}

// genericReplicaSetFinalizer implements a generic reconcile.FinalizingReconciler
type genericReplicaSetFinalizer struct {
	genericReplicaSetReconciler
	finalizingReconciler ReplicaSetFinalizer
}

func (r genericReplicaSetFinalizer) FinalizerName() string {
	return r.finalizingReconciler.ReplicaSetFinalizerName()
}

func (r genericReplicaSetFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*apps_v1.ReplicaSet)
	if !ok {
		return errors.Errorf("internal error: ReplicaSet handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeReplicaSet(obj)
}
