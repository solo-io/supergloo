// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	. "github.com/solo-io/mesh-projects/pkg/api/config.zephyr.solo.io/v1alpha1"

	"github.com/pkg/errors"
	"github.com/solo-io/autopilot/pkg/events"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type RoutingRuleEventHandler interface {
	Create(obj *RoutingRule) error
	Update(old, new *RoutingRule) error
	Delete(obj *RoutingRule) error
	Generic(obj *RoutingRule) error
}

type RoutingRuleEventHandlerFuncs struct {
	OnCreate  func(obj *RoutingRule) error
	OnUpdate  func(old, new *RoutingRule) error
	OnDelete  func(obj *RoutingRule) error
	OnGeneric func(obj *RoutingRule) error
}

func (f *RoutingRuleEventHandlerFuncs) Create(obj *RoutingRule) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *RoutingRuleEventHandlerFuncs) Delete(obj *RoutingRule) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *RoutingRuleEventHandlerFuncs) Update(objOld, objNew *RoutingRule) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *RoutingRuleEventHandlerFuncs) Generic(obj *RoutingRule) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type RoutingRuleController interface {
	AddEventHandler(ctx context.Context, h RoutingRuleEventHandler, predicates ...predicate.Predicate) error
}

type RoutingRuleControllerImpl struct {
	watcher events.EventWatcher
}

func NewRoutingRuleController(name string, mgr manager.Manager) (RoutingRuleController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &RoutingRuleControllerImpl{
		watcher: w,
	}, nil
}

func (c *RoutingRuleControllerImpl) AddEventHandler(ctx context.Context, h RoutingRuleEventHandler, predicates ...predicate.Predicate) error {
	handler := genericRoutingRuleHandler{handler: h}
	if err := c.watcher.Watch(ctx, &RoutingRule{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericRoutingRuleHandler implements a generic events.EventHandler
type genericRoutingRuleHandler struct {
	handler RoutingRuleEventHandler
}

func (h genericRoutingRuleHandler) Create(object runtime.Object) error {
	obj, ok := object.(*RoutingRule)
	if !ok {
		return errors.Errorf("internal error: RoutingRule handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericRoutingRuleHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*RoutingRule)
	if !ok {
		return errors.Errorf("internal error: RoutingRule handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericRoutingRuleHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*RoutingRule)
	if !ok {
		return errors.Errorf("internal error: RoutingRule handler received event for %T", old)
	}
	objNew, ok := new.(*RoutingRule)
	if !ok {
		return errors.Errorf("internal error: RoutingRule handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericRoutingRuleHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*RoutingRule)
	if !ok {
		return errors.Errorf("internal error: RoutingRule handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

type SecurityRuleEventHandler interface {
	Create(obj *SecurityRule) error
	Update(old, new *SecurityRule) error
	Delete(obj *SecurityRule) error
	Generic(obj *SecurityRule) error
}

type SecurityRuleEventHandlerFuncs struct {
	OnCreate  func(obj *SecurityRule) error
	OnUpdate  func(old, new *SecurityRule) error
	OnDelete  func(obj *SecurityRule) error
	OnGeneric func(obj *SecurityRule) error
}

func (f *SecurityRuleEventHandlerFuncs) Create(obj *SecurityRule) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *SecurityRuleEventHandlerFuncs) Delete(obj *SecurityRule) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *SecurityRuleEventHandlerFuncs) Update(objOld, objNew *SecurityRule) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *SecurityRuleEventHandlerFuncs) Generic(obj *SecurityRule) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type SecurityRuleController interface {
	AddEventHandler(ctx context.Context, h SecurityRuleEventHandler, predicates ...predicate.Predicate) error
}

type SecurityRuleControllerImpl struct {
	watcher events.EventWatcher
}

func NewSecurityRuleController(name string, mgr manager.Manager) (SecurityRuleController, error) {
	if err := AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}

	w, err := events.NewWatcher(name, mgr)
	if err != nil {
		return nil, err
	}
	return &SecurityRuleControllerImpl{
		watcher: w,
	}, nil
}

func (c *SecurityRuleControllerImpl) AddEventHandler(ctx context.Context, h SecurityRuleEventHandler, predicates ...predicate.Predicate) error {
	handler := genericSecurityRuleHandler{handler: h}
	if err := c.watcher.Watch(ctx, &SecurityRule{}, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericSecurityRuleHandler implements a generic events.EventHandler
type genericSecurityRuleHandler struct {
	handler SecurityRuleEventHandler
}

func (h genericSecurityRuleHandler) Create(object runtime.Object) error {
	obj, ok := object.(*SecurityRule)
	if !ok {
		return errors.Errorf("internal error: SecurityRule handler received event for %T", object)
	}
	return h.handler.Create(obj)
}

func (h genericSecurityRuleHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*SecurityRule)
	if !ok {
		return errors.Errorf("internal error: SecurityRule handler received event for %T", object)
	}
	return h.handler.Delete(obj)
}

func (h genericSecurityRuleHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*SecurityRule)
	if !ok {
		return errors.Errorf("internal error: SecurityRule handler received event for %T", old)
	}
	objNew, ok := new.(*SecurityRule)
	if !ok {
		return errors.Errorf("internal error: SecurityRule handler received event for %T", new)
	}
	return h.handler.Update(objOld, objNew)
}

func (h genericSecurityRuleHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*SecurityRule)
	if !ok {
		return errors.Errorf("internal error: SecurityRule handler received event for %T", object)
	}
	return h.handler.Generic(obj)
}

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
