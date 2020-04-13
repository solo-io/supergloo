// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	core_v1 "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Handle events for the Secret Resource
type SecretEventHandler interface {
	CreateSecret(obj *core_v1.Secret) error
	UpdateSecret(old, new *core_v1.Secret) error
	DeleteSecret(obj *core_v1.Secret) error
	GenericSecret(obj *core_v1.Secret) error
}

type SecretEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Secret) error
	OnUpdate  func(old, new *core_v1.Secret) error
	OnDelete  func(obj *core_v1.Secret) error
	OnGeneric func(obj *core_v1.Secret) error
}

func (f *SecretEventHandlerFuncs) CreateSecret(obj *core_v1.Secret) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *SecretEventHandlerFuncs) DeleteSecret(obj *core_v1.Secret) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *SecretEventHandlerFuncs) UpdateSecret(objOld, objNew *core_v1.Secret) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *SecretEventHandlerFuncs) GenericSecret(obj *core_v1.Secret) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type SecretEventWatcher interface {
	AddEventHandler(ctx context.Context, h SecretEventHandler, predicates ...predicate.Predicate) error
}

type secretEventWatcher struct {
	watcher events.EventWatcher
}

func NewSecretEventWatcher(name string, mgr manager.Manager) SecretEventWatcher {
	return &secretEventWatcher{
		watcher: events.NewWatcher(name, mgr, &core_v1.Secret{}),
	}
}

func (c *secretEventWatcher) AddEventHandler(ctx context.Context, h SecretEventHandler, predicates ...predicate.Predicate) error {
	handler := genericSecretHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
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
	return h.handler.CreateSecret(obj)
}

func (h genericSecretHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.DeleteSecret(obj)
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
	return h.handler.UpdateSecret(objOld, objNew)
}

func (h genericSecretHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return h.handler.GenericSecret(obj)
}

// Handle events for the Service Resource
type ServiceEventHandler interface {
	CreateService(obj *core_v1.Service) error
	UpdateService(old, new *core_v1.Service) error
	DeleteService(obj *core_v1.Service) error
	GenericService(obj *core_v1.Service) error
}

type ServiceEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Service) error
	OnUpdate  func(old, new *core_v1.Service) error
	OnDelete  func(obj *core_v1.Service) error
	OnGeneric func(obj *core_v1.Service) error
}

func (f *ServiceEventHandlerFuncs) CreateService(obj *core_v1.Service) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *ServiceEventHandlerFuncs) DeleteService(obj *core_v1.Service) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *ServiceEventHandlerFuncs) UpdateService(objOld, objNew *core_v1.Service) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *ServiceEventHandlerFuncs) GenericService(obj *core_v1.Service) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type ServiceEventWatcher interface {
	AddEventHandler(ctx context.Context, h ServiceEventHandler, predicates ...predicate.Predicate) error
}

type serviceEventWatcher struct {
	watcher events.EventWatcher
}

func NewServiceEventWatcher(name string, mgr manager.Manager) ServiceEventWatcher {
	return &serviceEventWatcher{
		watcher: events.NewWatcher(name, mgr, &core_v1.Service{}),
	}
}

func (c *serviceEventWatcher) AddEventHandler(ctx context.Context, h ServiceEventHandler, predicates ...predicate.Predicate) error {
	handler := genericServiceHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
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
	return h.handler.CreateService(obj)
}

func (h genericServiceHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.DeleteService(obj)
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
	return h.handler.UpdateService(objOld, objNew)
}

func (h genericServiceHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return h.handler.GenericService(obj)
}

// Handle events for the Pod Resource
type PodEventHandler interface {
	CreatePod(obj *core_v1.Pod) error
	UpdatePod(old, new *core_v1.Pod) error
	DeletePod(obj *core_v1.Pod) error
	GenericPod(obj *core_v1.Pod) error
}

type PodEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Pod) error
	OnUpdate  func(old, new *core_v1.Pod) error
	OnDelete  func(obj *core_v1.Pod) error
	OnGeneric func(obj *core_v1.Pod) error
}

func (f *PodEventHandlerFuncs) CreatePod(obj *core_v1.Pod) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *PodEventHandlerFuncs) DeletePod(obj *core_v1.Pod) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *PodEventHandlerFuncs) UpdatePod(objOld, objNew *core_v1.Pod) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *PodEventHandlerFuncs) GenericPod(obj *core_v1.Pod) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type PodEventWatcher interface {
	AddEventHandler(ctx context.Context, h PodEventHandler, predicates ...predicate.Predicate) error
}

type podEventWatcher struct {
	watcher events.EventWatcher
}

func NewPodEventWatcher(name string, mgr manager.Manager) PodEventWatcher {
	return &podEventWatcher{
		watcher: events.NewWatcher(name, mgr, &core_v1.Pod{}),
	}
}

func (c *podEventWatcher) AddEventHandler(ctx context.Context, h PodEventHandler, predicates ...predicate.Predicate) error {
	handler := genericPodHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
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
	return h.handler.CreatePod(obj)
}

func (h genericPodHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.DeletePod(obj)
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
	return h.handler.UpdatePod(objOld, objNew)
}

func (h genericPodHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return h.handler.GenericPod(obj)
}
