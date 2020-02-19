// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	core_v1 "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SecretEventHandler interface {
	Create(obj *core_v1.Secret) error
	Update(old, new *core_v1.Secret) error
	Delete(obj *core_v1.Secret) error
	Generic(obj *core_v1.Secret) error
}

type SecretEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Secret) error
	OnUpdate  func(old, new *core_v1.Secret) error
	OnDelete  func(obj *core_v1.Secret) error
	OnGeneric func(obj *core_v1.Secret) error
}

func (f *SecretEventHandlerFuncs) Create(obj *core_v1.Secret) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *SecretEventHandlerFuncs) Delete(obj *core_v1.Secret) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *SecretEventHandlerFuncs) Update(objOld, objNew *core_v1.Secret) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *SecretEventHandlerFuncs) Generic(obj *core_v1.Secret) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type SecretController interface {
	AddEventHandler(ctx context.Context, h SecretEventHandler, predicates ...predicate.Predicate) error
}

type SecretControllerImpl struct {
	watcher events.EventWatcher
}

func NewSecretController(name string, mgr manager.Manager) (SecretController, error) {
	if err := core_v1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &SecretControllerImpl{
		watcher: w,
	}, nil
}

func (c *SecretControllerImpl) AddEventHandler(ctx context.Context, h SecretEventHandler, predicates ...predicate.Predicate) error {
	handler := genericSecretHandler{handler: h}
	if err := c.watcher.Watch(ctx, &core_v1.Secret{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericSecretHandler implements a generic events.EventHandler
type genericSecretHandler struct {
	handler SecretEventHandler
}

func (h genericSecretHandler) Create(object runtime.Object) error {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericSecretHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericSecretHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", old)
	}
	objNew, ok := new.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericSecretHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type ServiceEventHandler interface {
	Create(obj *core_v1.Service) error
	Update(old, new *core_v1.Service) error
	Delete(obj *core_v1.Service) error
	Generic(obj *core_v1.Service) error
}

type ServiceEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Service) error
	OnUpdate  func(old, new *core_v1.Service) error
	OnDelete  func(obj *core_v1.Service) error
	OnGeneric func(obj *core_v1.Service) error
}

func (f *ServiceEventHandlerFuncs) Create(obj *core_v1.Service) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *ServiceEventHandlerFuncs) Delete(obj *core_v1.Service) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *ServiceEventHandlerFuncs) Update(objOld, objNew *core_v1.Service) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *ServiceEventHandlerFuncs) Generic(obj *core_v1.Service) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type ServiceController interface {
	AddEventHandler(ctx context.Context, h ServiceEventHandler, predicates ...predicate.Predicate) error
}

type ServiceControllerImpl struct {
	watcher events.EventWatcher
}

func NewServiceController(name string, mgr manager.Manager) (ServiceController, error) {
	if err := core_v1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &ServiceControllerImpl{
		watcher: w,
	}, nil
}

func (c *ServiceControllerImpl) AddEventHandler(ctx context.Context, h ServiceEventHandler, predicates ...predicate.Predicate) error {
	handler := genericServiceHandler{handler: h}
	if err := c.watcher.Watch(ctx, &core_v1.Service{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericServiceHandler implements a generic events.EventHandler
type genericServiceHandler struct {
	handler ServiceEventHandler
}

func (h genericServiceHandler) Create(object runtime.Object) error {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericServiceHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericServiceHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", old)
	}
	objNew, ok := new.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericServiceHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type PodEventHandler interface {
	Create(obj *core_v1.Pod) error
	Update(old, new *core_v1.Pod) error
	Delete(obj *core_v1.Pod) error
	Generic(obj *core_v1.Pod) error
}

type PodEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Pod) error
	OnUpdate  func(old, new *core_v1.Pod) error
	OnDelete  func(obj *core_v1.Pod) error
	OnGeneric func(obj *core_v1.Pod) error
}

func (f *PodEventHandlerFuncs) Create(obj *core_v1.Pod) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *PodEventHandlerFuncs) Delete(obj *core_v1.Pod) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *PodEventHandlerFuncs) Update(objOld, objNew *core_v1.Pod) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *PodEventHandlerFuncs) Generic(obj *core_v1.Pod) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type PodController interface {
	AddEventHandler(ctx context.Context, h PodEventHandler, predicates ...predicate.Predicate) error
}

type PodControllerImpl struct {
	watcher events.EventWatcher
}

func NewPodController(name string, mgr manager.Manager) (PodController, error) {
	if err := core_v1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &PodControllerImpl{
		watcher: w,
	}, nil
}

func (c *PodControllerImpl) AddEventHandler(ctx context.Context, h PodEventHandler, predicates ...predicate.Predicate) error {
	handler := genericPodHandler{handler: h}
	if err := c.watcher.Watch(ctx, &core_v1.Pod{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericPodHandler implements a generic events.EventHandler
type genericPodHandler struct {
	handler PodEventHandler
}

func (h genericPodHandler) Create(object runtime.Object) error {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericPodHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericPodHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", old)
	}
	objNew, ok := new.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericPodHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
