// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	. "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"

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

type KubernetesClusterController struct {
	watcher events.EventWatcher
}

func NewKubernetesClusterController(name string, mgr manager.Manager) (*KubernetesClusterController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &KubernetesClusterController{
		watcher: w,
	}, nil
}

func (c *KubernetesClusterController) AddEventHandler(ctx context.Context, h KubernetesClusterEventHandler, predicates ...predicate.Predicate) error {
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
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T")
	}
	return h.handler.Create(obj)
}

func (h genericKubernetesClusterHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T")
	}
	return h.handler.Delete(obj)
}

func (h genericKubernetesClusterHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T")
	}
	objNew, ok := new.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T")
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericKubernetesClusterHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*KubernetesCluster)
	if !ok {
		return errors.Errorf("internal error: KubernetesCluster handler received event for %T")
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

type ServiceController struct {
	watcher events.EventWatcher
}

func NewServiceController(name string, mgr manager.Manager) (*ServiceController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &ServiceController{
		watcher: w,
	}, nil
}

func (c *ServiceController) AddEventHandler(ctx context.Context, h ServiceEventHandler, predicates ...predicate.Predicate) error {
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
		return errors.Errorf("internal error: Service handler received event for %T")
	}
	return h.handler.Create(obj)
}

func (h genericServiceHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T")
	}
	return h.handler.Delete(obj)
}

func (h genericServiceHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T")
	}
	objNew, ok := new.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T")
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericServiceHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T")
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

type MeshController struct {
	watcher events.EventWatcher
}

func NewMeshController(name string, mgr manager.Manager) (*MeshController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &MeshController{
		watcher: w,
	}, nil
}

func (c *MeshController) AddEventHandler(ctx context.Context, h MeshEventHandler, predicates ...predicate.Predicate) error {
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
		return errors.Errorf("internal error: Mesh handler received event for %T")
	}
	return h.handler.Create(obj)
}

func (h genericMeshHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T")
	}
	return h.handler.Delete(obj)
}

func (h genericMeshHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T")
	}
	objNew, ok := new.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T")
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*Mesh)
	if !ok {
		return errors.Errorf("internal error: Mesh handler received event for %T")
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

type MeshGroupController struct {
	watcher events.EventWatcher
}

func NewMeshGroupController(name string, mgr manager.Manager) (*MeshGroupController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &MeshGroupController{
		watcher: w,
	}, nil
}

func (c *MeshGroupController) AddEventHandler(ctx context.Context, h MeshGroupEventHandler, predicates ...predicate.Predicate) error {
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
		return errors.Errorf("internal error: MeshGroup handler received event for %T")
	}
	return h.handler.Create(obj)
}

func (h genericMeshGroupHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T")
	}
	return h.handler.Delete(obj)
}

func (h genericMeshGroupHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T")
	}
	objNew, ok := new.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T")
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshGroupHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*MeshGroup)
	if !ok {
		return errors.Errorf("internal error: MeshGroup handler received event for %T")
	}
	return h.handler.Generic(obj)
}
