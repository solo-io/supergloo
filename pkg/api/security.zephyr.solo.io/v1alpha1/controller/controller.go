// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	security_zephyr_solo_io_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type MeshGroupCertificateSigningRequestEventHandler interface {
	Create(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
	Update(old, new *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
	Delete(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
	Generic(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
}

type MeshGroupCertificateSigningRequestEventHandlerFuncs struct {
	OnCreate  func(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
	OnUpdate  func(old, new *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
	OnDelete  func(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
	OnGeneric func(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Create(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Delete(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Update(objOld, objNew *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *MeshGroupCertificateSigningRequestEventHandlerFuncs) Generic(obj *security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest) error {
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
	if err := security_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
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
	if err := c.watcher.Watch(ctx, &security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericMeshGroupCertificateSigningRequestHandler implements a generic events.EventHandler
type genericMeshGroupCertificateSigningRequestHandler struct {
	handler MeshGroupCertificateSigningRequestEventHandler
}

func (h genericMeshGroupCertificateSigningRequestHandler) Create(object runtime.Object) error {
	obj, ok := object.(*security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericMeshGroupCertificateSigningRequestHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericMeshGroupCertificateSigningRequestHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", old)
	}
	objNew, ok := new.(*security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericMeshGroupCertificateSigningRequestHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*security_zephyr_solo_io_v1alpha1.MeshGroupCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: MeshGroupCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
