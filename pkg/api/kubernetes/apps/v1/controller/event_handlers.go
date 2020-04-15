// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	apps_v1 "k8s.io/api/apps/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Handle events for the Deployment Resource
type DeploymentEventHandler interface {
	CreateDeployment(obj *apps_v1.Deployment) error
	UpdateDeployment(old, new *apps_v1.Deployment) error
	DeleteDeployment(obj *apps_v1.Deployment) error
	GenericDeployment(obj *apps_v1.Deployment) error
}

type DeploymentEventHandlerFuncs struct {
	OnCreate  func(obj *apps_v1.Deployment) error
	OnUpdate  func(old, new *apps_v1.Deployment) error
	OnDelete  func(obj *apps_v1.Deployment) error
	OnGeneric func(obj *apps_v1.Deployment) error
}

func (f *DeploymentEventHandlerFuncs) CreateDeployment(obj *apps_v1.Deployment) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *DeploymentEventHandlerFuncs) DeleteDeployment(obj *apps_v1.Deployment) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *DeploymentEventHandlerFuncs) UpdateDeployment(objOld, objNew *apps_v1.Deployment) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *DeploymentEventHandlerFuncs) GenericDeployment(obj *apps_v1.Deployment) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type DeploymentEventWatcher interface {
	AddEventHandler(ctx context.Context, h DeploymentEventHandler, predicates ...predicate.Predicate) error
}

type deploymentEventWatcher struct {
	watcher events.EventWatcher
}

func NewDeploymentEventWatcher(name string, mgr manager.Manager) DeploymentEventWatcher {
	return &deploymentEventWatcher{
		watcher: events.NewWatcher(name, mgr, &apps_v1.Deployment{}),
	}
}

func (c *deploymentEventWatcher) AddEventHandler(ctx context.Context, h DeploymentEventHandler, predicates ...predicate.Predicate) error {
	handler := genericDeploymentHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
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
	return h.handler.CreateDeployment(obj)
}

func (h genericDeploymentHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", object)
	}
	return h.handler.DeleteDeployment(obj)
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
	return h.handler.UpdateDeployment(objOld, objNew)
}

func (h genericDeploymentHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*apps_v1.Deployment)
	if !ok {
		return errors.Errorf("internal error: Deployment handler received event for %T", object)
	}
	return h.handler.GenericDeployment(obj)
}

// Handle events for the ReplicaSet Resource
type ReplicaSetEventHandler interface {
	CreateReplicaSet(obj *apps_v1.ReplicaSet) error
	UpdateReplicaSet(old, new *apps_v1.ReplicaSet) error
	DeleteReplicaSet(obj *apps_v1.ReplicaSet) error
	GenericReplicaSet(obj *apps_v1.ReplicaSet) error
}

type ReplicaSetEventHandlerFuncs struct {
	OnCreate  func(obj *apps_v1.ReplicaSet) error
	OnUpdate  func(old, new *apps_v1.ReplicaSet) error
	OnDelete  func(obj *apps_v1.ReplicaSet) error
	OnGeneric func(obj *apps_v1.ReplicaSet) error
}

func (f *ReplicaSetEventHandlerFuncs) CreateReplicaSet(obj *apps_v1.ReplicaSet) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *ReplicaSetEventHandlerFuncs) DeleteReplicaSet(obj *apps_v1.ReplicaSet) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *ReplicaSetEventHandlerFuncs) UpdateReplicaSet(objOld, objNew *apps_v1.ReplicaSet) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *ReplicaSetEventHandlerFuncs) GenericReplicaSet(obj *apps_v1.ReplicaSet) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type ReplicaSetEventWatcher interface {
	AddEventHandler(ctx context.Context, h ReplicaSetEventHandler, predicates ...predicate.Predicate) error
}

type replicaSetEventWatcher struct {
	watcher events.EventWatcher
}

func NewReplicaSetEventWatcher(name string, mgr manager.Manager) ReplicaSetEventWatcher {
	return &replicaSetEventWatcher{
		watcher: events.NewWatcher(name, mgr, &apps_v1.ReplicaSet{}),
	}
}

func (c *replicaSetEventWatcher) AddEventHandler(ctx context.Context, h ReplicaSetEventHandler, predicates ...predicate.Predicate) error {
	handler := genericReplicaSetHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericReplicaSetHandler implements a generic events.EventHandler
type genericReplicaSetHandler struct {
	handler ReplicaSetEventHandler
}

func (h genericReplicaSetHandler) Create(object runtime.Object) error {
	obj, ok := object.(*apps_v1.ReplicaSet)
	if !ok {
		return errors.Errorf("internal error: ReplicaSet handler received event for %T", object)
	}
	return h.handler.CreateReplicaSet(obj)
}

func (h genericReplicaSetHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*apps_v1.ReplicaSet)
	if !ok {
		return errors.Errorf("internal error: ReplicaSet handler received event for %T", object)
	}
	return h.handler.DeleteReplicaSet(obj)
}

func (h genericReplicaSetHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*apps_v1.ReplicaSet)
	if !ok {
		return errors.Errorf("internal error: ReplicaSet handler received event for %T", old)
	}
	objNew, ok := new.(*apps_v1.ReplicaSet)
	if !ok {
		return errors.Errorf("internal error: ReplicaSet handler received event for %T", new)
	}
	return h.handler.UpdateReplicaSet(objOld, objNew)
}

func (h genericReplicaSetHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*apps_v1.ReplicaSet)
	if !ok {
		return errors.Errorf("internal error: ReplicaSet handler received event for %T", object)
	}
	return h.handler.GenericReplicaSet(obj)
}
