// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	. "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type MeshGroupCertificateSigningRequestEventHandler interface {
	Create(obj *MeshGroupCertificateSigningRequest) error
	Update(old, new *MeshGroupCertificateSigningRequest) error
	Delete(obj *MeshGroupCertificateSigningRequest) error
	Generic(obj *MeshGroupCertificateSigningRequest) error
}

type MeshGroupCertificateSigningRequestEventHandlerFuncs struct {
	OnCreate  func(obj *MeshGroupCertificateSigningRequest) error
	OnUpdate  func(old, new *MeshGroupCertificateSigningRequest) error
	OnDelete  func(obj *MeshGroupCertificateSigningRequest) error
	OnGeneric func(obj *MeshGroupCertificateSigningRequest) error
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Create(obj *MeshGroupCertificateSigningRequest) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Delete(obj *MeshGroupCertificateSigningRequest) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Update(objOld, objNew *MeshGroupCertificateSigningRequest) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Generic(obj *MeshGroupCertificateSigningRequest) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type MeshGroupCertificateSigningRequestController interface {
	AddEventHandler(ctx context.Context, h MeshGroupCertificateSigningRequestEventHandler, predicates ...predicate.Predicate) error
}

type MeshGroupCertificateSigningRequestControllerImpl struct {
	watcher events.EventWatcher
}

func NewMeshGroupCertificateSigningRequestController(name string, mgr manager.Manager) (MeshGroupCertificateSigningRequestController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &MeshGroupCertificateSigningRequestControllerImpl{
		watcher: w,
	}, nil
}

func (c *MeshGroupCertificateSigningRequestControllerImpl) AddEventHandler(ctx context.Context, h MeshGroupCertificateSigningRequestEventHandler, predicates ...predicate.Predicate) error {
	handler := genericMeshGroupCertificateSigningRequestHandler{handler: h}
	if err := c.watcher.Watch(ctx, &MeshGroupCertificateSigningRequest{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshGroupCertificateSigningRequestHandler implements a generic events.EventHandler
type genericMeshGroupCertificateSigningRequestHandler struct {
	handler MeshGroupCertificateSigningRequestEventHandler
}

func (h genericMeshGroupCertificateSigningRequestHandler) Create(object runtime.Object) error {
	obj, ok := object.(*MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshGroupCertificateSigningRequestHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshGroupCertificateSigningRequestHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", old)
	}
	objNew, ok := new.(*MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshGroupCertificateSigningRequestHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
