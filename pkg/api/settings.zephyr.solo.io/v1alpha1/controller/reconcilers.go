// Code generated by skv2. DO NOT EDIT.

// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	settings_zephyr_solo_io_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/settings.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconcile Upsert events for the Settings Resource.
// implemented by the user
type SettingsReconciler interface {
	ReconcileSettings(obj *settings_zephyr_solo_io_v1alpha1.Settings) (reconcile.Result, error)
}

// Reconcile deletion events for the Settings Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type SettingsDeletionReconciler interface {
	ReconcileSettingsDeletion(req reconcile.Request)
}

type SettingsReconcilerFuncs struct {
	OnReconcileSettings         func(obj *settings_zephyr_solo_io_v1alpha1.Settings) (reconcile.Result, error)
	OnReconcileSettingsDeletion func(req reconcile.Request)
}

func (f *SettingsReconcilerFuncs) ReconcileSettings(obj *settings_zephyr_solo_io_v1alpha1.Settings) (reconcile.Result, error) {
	if f.OnReconcileSettings == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileSettings(obj)
}

func (f *SettingsReconcilerFuncs) ReconcileSettingsDeletion(req reconcile.Request) {
	if f.OnReconcileSettingsDeletion == nil {
		return
	}
	f.OnReconcileSettingsDeletion(req)
}

// Reconcile and finalize the Settings Resource
// implemented by the user
type SettingsFinalizer interface {
	SettingsReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	SettingsFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeSettings(obj *settings_zephyr_solo_io_v1alpha1.Settings) error
}

type SettingsReconcileLoop interface {
	RunSettingsReconciler(ctx context.Context, rec SettingsReconciler, predicates ...predicate.Predicate) error
}

type settingsReconcileLoop struct {
	loop reconcile.Loop
}

func NewSettingsReconcileLoop(name string, mgr manager.Manager) SettingsReconcileLoop {
	return &settingsReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &settings_zephyr_solo_io_v1alpha1.Settings{}),
	}
}

func (c *settingsReconcileLoop) RunSettingsReconciler(ctx context.Context, reconciler SettingsReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericSettingsReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(SettingsFinalizer); ok {
		reconcilerWrapper = genericSettingsFinalizer{
			genericSettingsReconciler: genericReconciler,
			finalizingReconciler:      finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericSettingsHandler implements a generic reconcile.Reconciler
type genericSettingsReconciler struct {
	reconciler SettingsReconciler
}

func (r genericSettingsReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*settings_zephyr_solo_io_v1alpha1.Settings)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Settings handler received event for %T", object)
	}
	return r.reconciler.ReconcileSettings(obj)
}

func (r genericSettingsReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(SettingsDeletionReconciler); ok {
		deletionReconciler.ReconcileSettingsDeletion(request)
	}
}

// genericSettingsFinalizer implements a generic reconcile.FinalizingReconciler
type genericSettingsFinalizer struct {
	genericSettingsReconciler
	finalizingReconciler SettingsFinalizer
}

func (r genericSettingsFinalizer) FinalizerName() string {
	return r.finalizingReconciler.SettingsFinalizerName()
}

func (r genericSettingsFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*settings_zephyr_solo_io_v1alpha1.Settings)
	if !ok {
		return errors.Errorf("internal error: Settings handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeSettings(obj)
}
