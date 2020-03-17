// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	networking_zephyr_solo_io_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type TrafficPolicyEventHandler interface {
	Create(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
	Update(old, new *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
	Delete(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
	Generic(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
}

type TrafficPolicyEventHandlerFuncs struct {
	OnCreate  func(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
	OnUpdate  func(old, new *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
	OnDelete  func(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
	OnGeneric func(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error
}

func (f *TrafficPolicyEventHandlerFuncs) Create(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *TrafficPolicyEventHandlerFuncs) Delete(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *TrafficPolicyEventHandlerFuncs) Update(objOld, objNew *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *TrafficPolicyEventHandlerFuncs) Generic(obj *networking_zephyr_solo_io_v1alpha1.TrafficPolicy) error {
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
	if err := networking_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &networking_zephyr_solo_io_v1alpha1.TrafficPolicy{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericTrafficPolicyHandler implements a generic events.EventHandler
type genericTrafficPolicyHandler struct {
	handler TrafficPolicyEventHandler
}

func (h genericTrafficPolicyHandler) Create(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericTrafficPolicyHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericTrafficPolicyHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*networking_zephyr_solo_io_v1alpha1.TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", old)
	}
	objNew, ok := new.(*networking_zephyr_solo_io_v1alpha1.TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericTrafficPolicyHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.TrafficPolicy)
	if !ok {
		return errors.Errorf("internal error: TrafficPolicy handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type AccessControlPolicyEventHandler interface {
	Create(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
	Update(old, new *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
	Delete(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
	Generic(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
}

type AccessControlPolicyEventHandlerFuncs struct {
	OnCreate  func(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
	OnUpdate  func(old, new *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
	OnDelete  func(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
	OnGeneric func(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error
}

func (f *AccessControlPolicyEventHandlerFuncs) Create(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *AccessControlPolicyEventHandlerFuncs) Delete(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *AccessControlPolicyEventHandlerFuncs) Update(objOld, objNew *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *AccessControlPolicyEventHandlerFuncs) Generic(obj *networking_zephyr_solo_io_v1alpha1.AccessControlPolicy) error {
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
	if err := networking_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &networking_zephyr_solo_io_v1alpha1.AccessControlPolicy{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericAccessControlPolicyHandler implements a generic events.EventHandler
type genericAccessControlPolicyHandler struct {
	handler AccessControlPolicyEventHandler
}

func (h genericAccessControlPolicyHandler) Create(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericAccessControlPolicyHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericAccessControlPolicyHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*networking_zephyr_solo_io_v1alpha1.AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", old)
	}
	objNew, ok := new.(*networking_zephyr_solo_io_v1alpha1.AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericAccessControlPolicyHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.AccessControlPolicy)
	if !ok {
		return errors.Errorf("internal error: AccessControlPolicy handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type VirtualMeshEventHandler interface {
	Create(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
	Update(old, new *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
	Delete(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
	Generic(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
}

type VirtualMeshEventHandlerFuncs struct {
	OnCreate  func(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
	OnUpdate  func(old, new *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
	OnDelete  func(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
	OnGeneric func(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error
}

func (f *VirtualMeshEventHandlerFuncs) Create(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *VirtualMeshEventHandlerFuncs) Delete(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *VirtualMeshEventHandlerFuncs) Update(objOld, objNew *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *VirtualMeshEventHandlerFuncs) Generic(obj *networking_zephyr_solo_io_v1alpha1.VirtualMesh) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type VirtualMeshController interface {
	AddEventHandler(ctx context.Context, h VirtualMeshEventHandler, predicates ...predicate.Predicate) error
}

type VirtualMeshControllerImpl struct {
	watcher events.EventWatcher
}

func NewVirtualMeshController(name string, mgr manager.Manager) (VirtualMeshController, error) {
	if err := networking_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &VirtualMeshControllerImpl{
		watcher: w,
	}, nil
}

func (c *VirtualMeshControllerImpl) AddEventHandler(ctx context.Context, h VirtualMeshEventHandler, predicates ...predicate.Predicate) error {
	handler := genericVirtualMeshHandler{handler: h}
	if err := c.watcher.Watch(ctx, &networking_zephyr_solo_io_v1alpha1.VirtualMesh{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericVirtualMeshHandler implements a generic events.EventHandler
type genericVirtualMeshHandler struct {
	handler VirtualMeshEventHandler
}

func (h genericVirtualMeshHandler) Create(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.VirtualMesh)
	if !ok {
		return errors.Errorf("internal error: VirtualMesh handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericVirtualMeshHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.VirtualMesh)
	if !ok {
		return errors.Errorf("internal error: VirtualMesh handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericVirtualMeshHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*networking_zephyr_solo_io_v1alpha1.VirtualMesh)
	if !ok {
		return errors.Errorf("internal error: VirtualMesh handler received event for %T", old)
	}
	objNew, ok := new.(*networking_zephyr_solo_io_v1alpha1.VirtualMesh)
	if !ok {
		return errors.Errorf("internal error: VirtualMesh handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericVirtualMeshHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*networking_zephyr_solo_io_v1alpha1.VirtualMesh)
	if !ok {
		return errors.Errorf("internal error: VirtualMesh handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
