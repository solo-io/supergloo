// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	. "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type KubernetesClusterEventHandler interface {
	Create(obj *KubernetesCluster) error
	Update(old, new *KubernetesCluster) error
	Delete(obj *KubernetesCluster) error
	Generic(obj *KubernetesCluster) error
}

type KubernetesClusterEventHandlerFuncs struct {
	OnCreate  func(obj *KubernetesCluster) error
	OnUpdate  func(old, new *KubernetesCluster) error
	OnDelete  func(obj *KubernetesCluster) error
	OnGeneric func(obj *KubernetesCluster) error
}

func (f *KubernetesClusterEventHandlerFuncs) Create(obj *KubernetesCluster) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *KubernetesClusterEventHandlerFuncs) Delete(obj *KubernetesCluster) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *KubernetesClusterEventHandlerFuncs) Update(objOld, objNew *KubernetesCluster) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *KubernetesClusterEventHandlerFuncs) Generic(obj *KubernetesCluster) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type KubernetesClusterController interface {
	AddEventHandler(ctx context.Context, h KubernetesClusterEventHandler, predicates ...predicate.Predicate) error
}

type KubernetesClusterControllerImpl struct {
	watcher events.EventWatcher
}

func NewKubernetesClusterController(name string, mgr manager.Manager) (KubernetesClusterController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &KubernetesClusterControllerImpl{
		watcher: w,
	}, nil
}

func (c *KubernetesClusterControllerImpl) AddEventHandler(ctx context.Context, h KubernetesClusterEventHandler, predicates ...predicate.Predicate) error {
	handler := genericKubernetesClusterHandler{handler: h}
	if err := c.watcher.Watch(ctx, &KubernetesCluster{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericKubernetesClusterHandler implements a generic events.EventHandler
type genericKubernetesClusterHandler struct {
	handler KubernetesClusterEventHandler
}

func (h genericKubernetesClusterHandler) Create(object runtime.Object) error {
	obj, ok := object.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericKubernetesClusterHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericKubernetesClusterHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", old)
	}
	objNew, ok := new.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericKubernetesClusterHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type MeshServiceEventHandler interface {
	Create(obj *MeshService) error
	Update(old, new *MeshService) error
	Delete(obj *MeshService) error
	Generic(obj *MeshService) error
}

type MeshServiceEventHandlerFuncs struct {
	OnCreate  func(obj *MeshService) error
	OnUpdate  func(old, new *MeshService) error
	OnDelete  func(obj *MeshService) error
	OnGeneric func(obj *MeshService) error
}

func (f *MeshServiceEventHandlerFuncs) Create(obj *MeshService) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshServiceEventHandlerFuncs) Delete(obj *MeshService) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshServiceEventHandlerFuncs) Update(objOld, objNew *MeshService) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshServiceEventHandlerFuncs) Generic(obj *MeshService) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type MeshServiceController interface {
	AddEventHandler(ctx context.Context, h MeshServiceEventHandler, predicates ...predicate.Predicate) error
}

type MeshServiceControllerImpl struct {
	watcher events.EventWatcher
}

func NewMeshServiceController(name string, mgr manager.Manager) (MeshServiceController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &MeshServiceControllerImpl{
		watcher: w,
	}, nil
}

func (c *MeshServiceControllerImpl) AddEventHandler(ctx context.Context, h MeshServiceEventHandler, predicates ...predicate.Predicate) error {
	handler := genericMeshServiceHandler{handler: h}
	if err := c.watcher.Watch(ctx, &MeshService{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshServiceHandler implements a generic events.EventHandler
type genericMeshServiceHandler struct {
	handler MeshServiceEventHandler
}

func (h genericMeshServiceHandler) Create(object runtime.Object) error {
	obj, ok := object.(*MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshServiceHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshServiceHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", old)
	}
	objNew, ok := new.(*MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshServiceHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type MeshWorkloadEventHandler interface {
	Create(obj *MeshWorkload) error
	Update(old, new *MeshWorkload) error
	Delete(obj *MeshWorkload) error
	Generic(obj *MeshWorkload) error
}

type MeshWorkloadEventHandlerFuncs struct {
	OnCreate  func(obj *MeshWorkload) error
	OnUpdate  func(old, new *MeshWorkload) error
	OnDelete  func(obj *MeshWorkload) error
	OnGeneric func(obj *MeshWorkload) error
}

func (f *MeshWorkloadEventHandlerFuncs) Create(obj *MeshWorkload) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshWorkloadEventHandlerFuncs) Delete(obj *MeshWorkload) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshWorkloadEventHandlerFuncs) Update(objOld, objNew *MeshWorkload) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshWorkloadEventHandlerFuncs) Generic(obj *MeshWorkload) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type MeshWorkloadController interface {
	AddEventHandler(ctx context.Context, h MeshWorkloadEventHandler, predicates ...predicate.Predicate) error
}

type MeshWorkloadControllerImpl struct {
	watcher events.EventWatcher
}

func NewMeshWorkloadController(name string, mgr manager.Manager) (MeshWorkloadController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &MeshWorkloadControllerImpl{
		watcher: w,
	}, nil
}

func (c *MeshWorkloadControllerImpl) AddEventHandler(ctx context.Context, h MeshWorkloadEventHandler, predicates ...predicate.Predicate) error {
	handler := genericMeshWorkloadHandler{handler: h}
	if err := c.watcher.Watch(ctx, &MeshWorkload{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshWorkloadHandler implements a generic events.EventHandler
type genericMeshWorkloadHandler struct {
	handler MeshWorkloadEventHandler
}

func (h genericMeshWorkloadHandler) Create(object runtime.Object) error {
	obj, ok := object.(*MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshWorkloadHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshWorkloadHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", old)
	}
	objNew, ok := new.(*MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshWorkloadHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type MeshEventHandler interface {
	Create(obj *Mesh) error
	Update(old, new *Mesh) error
	Delete(obj *Mesh) error
	Generic(obj *Mesh) error
}

type MeshEventHandlerFuncs struct {
	OnCreate  func(obj *Mesh) error
	OnUpdate  func(old, new *Mesh) error
	OnDelete  func(obj *Mesh) error
	OnGeneric func(obj *Mesh) error
}

func (f *MeshEventHandlerFuncs) Create(obj *Mesh) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshEventHandlerFuncs) Delete(obj *Mesh) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshEventHandlerFuncs) Update(objOld, objNew *Mesh) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshEventHandlerFuncs) Generic(obj *Mesh) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type MeshController interface {
	AddEventHandler(ctx context.Context, h MeshEventHandler, predicates ...predicate.Predicate) error
}

type MeshControllerImpl struct {
	watcher events.EventWatcher
}

func NewMeshController(name string, mgr manager.Manager) (MeshController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &MeshControllerImpl{
		watcher: w,
	}, nil
}

func (c *MeshControllerImpl) AddEventHandler(ctx context.Context, h MeshEventHandler, predicates ...predicate.Predicate) error {
	handler := genericMeshHandler{handler: h}
	if err := c.watcher.Watch(ctx, &Mesh{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshHandler implements a generic events.EventHandler
type genericMeshHandler struct {
	handler MeshEventHandler
}

func (h genericMeshHandler) Create(object runtime.Object) error {
	obj, ok := object.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", old)
	}
	objNew, ok := new.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
