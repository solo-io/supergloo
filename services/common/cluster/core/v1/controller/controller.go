// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	. "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SecretEventHandler interface {
	Create(obj *Secret) error
	Update(old, new *Secret) error
	Delete(obj *Secret) error
	Generic(obj *Secret) error
}

type SecretEventHandlerFuncs struct {
	OnCreate  func(obj *Secret) error
	OnUpdate  func(old, new *Secret) error
	OnDelete  func(obj *Secret) error
	OnGeneric func(obj *Secret) error
}

func (f *SecretEventHandlerFuncs) Create(obj *Secret) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *SecretEventHandlerFuncs) Delete(obj *Secret) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *SecretEventHandlerFuncs) Update(objOld, objNew *Secret) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *SecretEventHandlerFuncs) Generic(obj *Secret) error {
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
	if err := AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &Secret{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericSecretHandler implements a generic events.EventHandler
type genericSecretHandler struct {
	handler SecretEventHandler
}

func (h genericSecretHandler) Create(object runtime.Object) error {
	obj, ok := object.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericSecretHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericSecretHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", old)
	}
	objNew, ok := new.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericSecretHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type ServiceEventHandler interface {
	Create(obj *Service) error
	Update(old, new *Service) error
	Delete(obj *Service) error
	Generic(obj *Service) error
}

type ServiceEventHandlerFuncs struct {
	OnCreate  func(obj *Service) error
	OnUpdate  func(old, new *Service) error
	OnDelete  func(obj *Service) error
	OnGeneric func(obj *Service) error
}

func (f *ServiceEventHandlerFuncs) Create(obj *Service) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *ServiceEventHandlerFuncs) Delete(obj *Service) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *ServiceEventHandlerFuncs) Update(objOld, objNew *Service) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *ServiceEventHandlerFuncs) Generic(obj *Service) error {
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
	if err := AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &Service{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericServiceHandler implements a generic events.EventHandler
type genericServiceHandler struct {
	handler ServiceEventHandler
}

func (h genericServiceHandler) Create(object runtime.Object) error {
	obj, ok := object.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericServiceHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericServiceHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", old)
	}
	objNew, ok := new.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericServiceHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type PodEventHandler interface {
	Create(obj *Pod) error
	Update(old, new *Pod) error
	Delete(obj *Pod) error
	Generic(obj *Pod) error
}

type PodEventHandlerFuncs struct {
	OnCreate  func(obj *Pod) error
	OnUpdate  func(old, new *Pod) error
	OnDelete  func(obj *Pod) error
	OnGeneric func(obj *Pod) error
}

func (f *PodEventHandlerFuncs) Create(obj *Pod) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *PodEventHandlerFuncs) Delete(obj *Pod) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *PodEventHandlerFuncs) Update(objOld, objNew *Pod) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *PodEventHandlerFuncs) Generic(obj *Pod) error {
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
	if err := AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &Pod{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericPodHandler implements a generic events.EventHandler
type genericPodHandler struct {
	handler PodEventHandler
}

func (h genericPodHandler) Create(object runtime.Object) error {
	obj, ok := object.(*Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericPodHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericPodHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", old)
	}
	objNew, ok := new.(*Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericPodHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
