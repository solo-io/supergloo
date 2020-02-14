// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	. "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type TrafficPolicyEventHandler interface {
	Create(obj *TrafficPolicy) error
	Update(old, new *TrafficPolicy) error
	Delete(obj *TrafficPolicy) error
	Generic(obj *TrafficPolicy) error
}

type TrafficPolicyEventHandlerFuncs struct {
	OnCreate  func(obj *TrafficPolicy) error
	OnUpdate  func(old, new *TrafficPolicy) error
	OnDelete  func(obj *TrafficPolicy) error
	OnGeneric func(obj *TrafficPolicy) error
}

func (f *TrafficPolicyEventHandlerFuncs) Create(obj *TrafficPolicy) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *TrafficPolicyEventHandlerFuncs) Delete(obj *TrafficPolicy) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *TrafficPolicyEventHandlerFuncs) Update(objOld, objNew *TrafficPolicy) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *TrafficPolicyEventHandlerFuncs) Generic(obj *TrafficPolicy) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type TrafficPolicyController interface {
	AddEventHandler(ctx context.Context, h TrafficPolicyEventHandler, predicates ...predicate.Predicate) error
}

type TrafficPolicyControllerImpl struct {
	watcher events.EventWatcher
}

func NewTrafficPolicyController(name string, mgr manager.Manager) (TrafficPolicyController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &TrafficPolicyControllerImpl{
		watcher: w,
	}, nil
}

func (c *TrafficPolicyControllerImpl) AddEventHandler(ctx context.Context, h TrafficPolicyEventHandler, predicates ...predicate.Predicate) error {
	handler := genericTrafficPolicyHandler{handler: h}
	if err := c.watcher.Watch(ctx, &TrafficPolicy{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericTrafficPolicyHandler implements a generic events.EventHandler
type genericTrafficPolicyHandler struct {
	handler TrafficPolicyEventHandler
}

func (h genericTrafficPolicyHandler) Create(object runtime.Object) error {
	obj, ok := object.(*TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericTrafficPolicyHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericTrafficPolicyHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", old)
	}
	objNew, ok := new.(*TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericTrafficPolicyHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type AccessControlPolicyEventHandler interface {
	Create(obj *AccessControlPolicy) error
	Update(old, new *AccessControlPolicy) error
	Delete(obj *AccessControlPolicy) error
	Generic(obj *AccessControlPolicy) error
}

type AccessControlPolicyEventHandlerFuncs struct {
	OnCreate  func(obj *AccessControlPolicy) error
	OnUpdate  func(old, new *AccessControlPolicy) error
	OnDelete  func(obj *AccessControlPolicy) error
	OnGeneric func(obj *AccessControlPolicy) error
}

func (f *AccessControlPolicyEventHandlerFuncs) Create(obj *AccessControlPolicy) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *AccessControlPolicyEventHandlerFuncs) Delete(obj *AccessControlPolicy) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *AccessControlPolicyEventHandlerFuncs) Update(objOld, objNew *AccessControlPolicy) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *AccessControlPolicyEventHandlerFuncs) Generic(obj *AccessControlPolicy) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type AccessControlPolicyController interface {
	AddEventHandler(ctx context.Context, h AccessControlPolicyEventHandler, predicates ...predicate.Predicate) error
}

type AccessControlPolicyControllerImpl struct {
	watcher events.EventWatcher
}

func NewAccessControlPolicyController(name string, mgr manager.Manager) (AccessControlPolicyController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &AccessControlPolicyControllerImpl{
		watcher: w,
	}, nil
}

func (c *AccessControlPolicyControllerImpl) AddEventHandler(ctx context.Context, h AccessControlPolicyEventHandler, predicates ...predicate.Predicate) error {
	handler := genericAccessControlPolicyHandler{handler: h}
	if err := c.watcher.Watch(ctx, &AccessControlPolicy{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericAccessControlPolicyHandler implements a generic events.EventHandler
type genericAccessControlPolicyHandler struct {
	handler AccessControlPolicyEventHandler
}

func (h genericAccessControlPolicyHandler) Create(object runtime.Object) error {
	obj, ok := object.(*AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericAccessControlPolicyHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericAccessControlPolicyHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", old)
	}
	objNew, ok := new.(*AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericAccessControlPolicyHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type MeshGroupEventHandler interface {
	Create(obj *MeshGroup) error
	Update(old, new *MeshGroup) error
	Delete(obj *MeshGroup) error
	Generic(obj *MeshGroup) error
}

type MeshGroupEventHandlerFuncs struct {
	OnCreate  func(obj *MeshGroup) error
	OnUpdate  func(old, new *MeshGroup) error
	OnDelete  func(obj *MeshGroup) error
	OnGeneric func(obj *MeshGroup) error
}

func (f *MeshGroupEventHandlerFuncs) Create(obj *MeshGroup) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshGroupEventHandlerFuncs) Delete(obj *MeshGroup) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshGroupEventHandlerFuncs) Update(objOld, objNew *MeshGroup) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshGroupEventHandlerFuncs) Generic(obj *MeshGroup) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type MeshGroupController interface {
	AddEventHandler(ctx context.Context, h MeshGroupEventHandler, predicates ...predicate.Predicate) error
}

type MeshGroupControllerImpl struct {
	watcher events.EventWatcher
}

func NewMeshGroupController(name string, mgr manager.Manager) (MeshGroupController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &MeshGroupControllerImpl{
		watcher: w,
	}, nil
}

func (c *MeshGroupControllerImpl) AddEventHandler(ctx context.Context, h MeshGroupEventHandler, predicates ...predicate.Predicate) error {
	handler := genericMeshGroupHandler{handler: h}
	if err := c.watcher.Watch(ctx, &MeshGroup{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshGroupHandler implements a generic events.EventHandler
type genericMeshGroupHandler struct {
	handler MeshGroupEventHandler
}

func (h genericMeshGroupHandler) Create(object runtime.Object) error {
	obj, ok := object.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshGroupHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshGroupHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T", old)
	}
	objNew, ok := new.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshGroupHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
