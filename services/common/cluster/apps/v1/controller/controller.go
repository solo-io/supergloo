// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	apps_v1 "k8s.io/api/apps/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type DeploymentEventHandler interface {
	Create(obj *apps_v1.Deployment) error
	Update(old, new *apps_v1.Deployment) error
	Delete(obj *apps_v1.Deployment) error
	Generic(obj *apps_v1.Deployment) error
}

type DeploymentEventHandlerFuncs struct {
	OnCreate  func(obj *apps_v1.Deployment) error
	OnUpdate  func(old, new *apps_v1.Deployment) error
	OnDelete  func(obj *apps_v1.Deployment) error
	OnGeneric func(obj *apps_v1.Deployment) error
}

func (f *DeploymentEventHandlerFuncs) Create(obj *apps_v1.Deployment) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *DeploymentEventHandlerFuncs) Delete(obj *apps_v1.Deployment) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *DeploymentEventHandlerFuncs) Update(objOld, objNew *apps_v1.Deployment) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *DeploymentEventHandlerFuncs) Generic(obj *apps_v1.Deployment) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type DeploymentController interface {
	AddEventHandler(ctx context.Context, h DeploymentEventHandler, predicates ...predicate.Predicate) error
}

type DeploymentControllerImpl struct {
	watcher events.EventWatcher
}

func NewDeploymentController(name string, mgr manager.Manager) (DeploymentController, error) {
	if err := apps_v1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &DeploymentControllerImpl{
		watcher: w,
	}, nil
}

func (c *DeploymentControllerImpl) AddEventHandler(ctx context.Context, h DeploymentEventHandler, predicates ...predicate.Predicate) error {
	handler := genericDeploymentHandler{handler: h}
	if err := c.watcher.Watch(ctx, &apps_v1.Deployment{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericDeploymentHandler implements a generic events.EventHandler
type genericDeploymentHandler struct {
	handler DeploymentEventHandler
}

func (h genericDeploymentHandler) Create(object runtime.Object) error {
	obj, ok := object.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericDeploymentHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericDeploymentHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", old)
	}
	objNew, ok := new.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericDeploymentHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
