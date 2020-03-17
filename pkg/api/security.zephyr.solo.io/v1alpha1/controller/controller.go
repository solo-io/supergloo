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

type VirtualMeshCertificateSigningRequestEventHandler interface {
	Create(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
	Update(old, new *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
	Delete(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
	Generic(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
}

type VirtualMeshCertificateSigningRequestEventHandlerFuncs struct {
	OnCreate  func(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
	OnUpdate  func(old, new *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
	OnDelete  func(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
	OnGeneric func(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error
}

func (f *VirtualMeshCertificateSigningRequestEventHandlerFuncs) Create(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *VirtualMeshCertificateSigningRequestEventHandlerFuncs) Delete(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *VirtualMeshCertificateSigningRequestEventHandlerFuncs) Update(objOld, objNew *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *VirtualMeshCertificateSigningRequestEventHandlerFuncs) Generic(obj *security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type VirtualMeshCertificateSigningRequestController interface {
	AddEventHandler(ctx context.Context, h VirtualMeshCertificateSigningRequestEventHandler, predicates ...predicate.Predicate) error
}

type VirtualMeshCertificateSigningRequestControllerImpl struct {
	watcher events.EventWatcher
}

func NewVirtualMeshCertificateSigningRequestController(name string, mgr manager.Manager) (VirtualMeshCertificateSigningRequestController, error) {
	if err := security_zephyr_solo_io_v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &VirtualMeshCertificateSigningRequestControllerImpl{
		watcher: w,
	}, nil
}

func (c *VirtualMeshCertificateSigningRequestControllerImpl) AddEventHandler(ctx context.Context, h VirtualMeshCertificateSigningRequestEventHandler, predicates ...predicate.Predicate) error {
	handler := genericVirtualMeshCertificateSigningRequestHandler{handler: h}
	if err := c.watcher.Watch(ctx, &security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericVirtualMeshCertificateSigningRequestHandler implements a generic events.EventHandler
type genericVirtualMeshCertificateSigningRequestHandler struct {
	handler VirtualMeshCertificateSigningRequestEventHandler
}

func (h genericVirtualMeshCertificateSigningRequestHandler) Create(object runtime.Object) error {
	obj, ok := object.(*security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: VirtualMeshCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericVirtualMeshCertificateSigningRequestHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: VirtualMeshCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericVirtualMeshCertificateSigningRequestHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: VirtualMeshCertificateSigningRequest handler received event for %T", old)
	}
	objNew, ok := new.(*security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: VirtualMeshCertificateSigningRequest handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericVirtualMeshCertificateSigningRequestHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*security_zephyr_solo_io_v1alpha1.VirtualMeshCertificateSigningRequest)
	if !ok {
		return errors.Errorf("internal error: VirtualMeshCertificateSigningRequest handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}
