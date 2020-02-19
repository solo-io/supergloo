// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	discovery_zephyr_solo_io_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type KubernetesClusterEventHandler interface {
	Create(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
	Update(old, new *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
	Delete(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
	Generic(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
}

type KubernetesClusterEventHandlerFuncs struct {
	OnCreate  func(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
	OnUpdate  func(old, new *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
	OnDelete  func(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
	OnGeneric func(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error
}

func (f *KubernetesClusterEventHandlerFuncs) Create(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *KubernetesClusterEventHandlerFuncs) Delete(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *KubernetesClusterEventHandlerFuncs) Update(objOld, objNew *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *KubernetesClusterEventHandlerFuncs) Generic(obj *discovery_zephyr_solo_io_v1alpha1.KubernetesCluster) error {
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
	if err := discovery_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &discovery_zephyr_solo_io_v1alpha1.KubernetesCluster{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericKubernetesClusterHandler implements a generic events.EventHandler
type genericKubernetesClusterHandler struct {
	handler KubernetesClusterEventHandler
}

func (h genericKubernetesClusterHandler) Create(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericKubernetesClusterHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericKubernetesClusterHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*discovery_zephyr_solo_io_v1alpha1.KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", old)
	}
	objNew, ok := new.(*discovery_zephyr_solo_io_v1alpha1.KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericKubernetesClusterHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type MeshServiceEventHandler interface {
	Create(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error
	Update(old, new *discovery_zephyr_solo_io_v1alpha1.MeshService) error
	Delete(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error
	Generic(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error
}

type MeshServiceEventHandlerFuncs struct {
	OnCreate  func(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error
	OnUpdate  func(old, new *discovery_zephyr_solo_io_v1alpha1.MeshService) error
	OnDelete  func(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error
	OnGeneric func(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error
}

func (f *MeshServiceEventHandlerFuncs) Create(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshServiceEventHandlerFuncs) Delete(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshServiceEventHandlerFuncs) Update(objOld, objNew *discovery_zephyr_solo_io_v1alpha1.MeshService) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshServiceEventHandlerFuncs) Generic(obj *discovery_zephyr_solo_io_v1alpha1.MeshService) error {
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
	if err := discovery_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &discovery_zephyr_solo_io_v1alpha1.MeshService{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshServiceHandler implements a generic events.EventHandler
type genericMeshServiceHandler struct {
	handler MeshServiceEventHandler
}

func (h genericMeshServiceHandler) Create(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshServiceHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshServiceHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*discovery_zephyr_solo_io_v1alpha1.MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", old)
	}
	objNew, ok := new.(*discovery_zephyr_solo_io_v1alpha1.MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshServiceHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.MeshService)
	if !ok {
		return errors.Errorf("internal error: MeshService handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type MeshWorkloadEventHandler interface {
	Create(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
	Update(old, new *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
	Delete(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
	Generic(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
}

type MeshWorkloadEventHandlerFuncs struct {
	OnCreate  func(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
	OnUpdate  func(old, new *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
	OnDelete  func(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
	OnGeneric func(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error
}

func (f *MeshWorkloadEventHandlerFuncs) Create(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshWorkloadEventHandlerFuncs) Delete(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshWorkloadEventHandlerFuncs) Update(objOld, objNew *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshWorkloadEventHandlerFuncs) Generic(obj *discovery_zephyr_solo_io_v1alpha1.MeshWorkload) error {
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
	if err := discovery_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &discovery_zephyr_solo_io_v1alpha1.MeshWorkload{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshWorkloadHandler implements a generic events.EventHandler
type genericMeshWorkloadHandler struct {
	handler MeshWorkloadEventHandler
}

func (h genericMeshWorkloadHandler) Create(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshWorkloadHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshWorkloadHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*discovery_zephyr_solo_io_v1alpha1.MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", old)
	}
	objNew, ok := new.(*discovery_zephyr_solo_io_v1alpha1.MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshWorkloadHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.MeshWorkload)
	if !ok {
		return errors.Errorf("internal error: MeshWorkload handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type MeshEventHandler interface {
	Create(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error
	Update(old, new *discovery_zephyr_solo_io_v1alpha1.Mesh) error
	Delete(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error
	Generic(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error
}

type MeshEventHandlerFuncs struct {
	OnCreate  func(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error
	OnUpdate  func(old, new *discovery_zephyr_solo_io_v1alpha1.Mesh) error
	OnDelete  func(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error
	OnGeneric func(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error
}

func (f *MeshEventHandlerFuncs) Create(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshEventHandlerFuncs) Delete(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshEventHandlerFuncs) Update(objOld, objNew *discovery_zephyr_solo_io_v1alpha1.Mesh) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshEventHandlerFuncs) Generic(obj *discovery_zephyr_solo_io_v1alpha1.Mesh) error {
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
	if err := discovery_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &discovery_zephyr_solo_io_v1alpha1.Mesh{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshHandler implements a generic events.EventHandler
type genericMeshHandler struct {
	handler MeshEventHandler
}

func (h genericMeshHandler) Create(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*discovery_zephyr_solo_io_v1alpha1.Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", old)
	}
	objNew, ok := new.(*discovery_zephyr_solo_io_v1alpha1.Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*discovery_zephyr_solo_io_v1alpha1.Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
