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

// Handle events for the ServiceAccount Resource
type ServiceAccountEventHandler interface {
	CreateServiceAccount(obj *core_v1.ServiceAccount) error
	UpdateServiceAccount(old, new *core_v1.ServiceAccount) error
	DeleteServiceAccount(obj *core_v1.ServiceAccount) error
	GenericServiceAccount(obj *core_v1.ServiceAccount) error
}

type ServiceAccountEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.ServiceAccount) error
	OnUpdate  func(old, new *core_v1.ServiceAccount) error
	OnDelete  func(obj *core_v1.ServiceAccount) error
	OnGeneric func(obj *core_v1.ServiceAccount) error
}

func (f *ServiceAccountEventHandlerFuncs) CreateServiceAccount(obj *core_v1.ServiceAccount) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *ServiceAccountEventHandlerFuncs) DeleteServiceAccount(obj *core_v1.ServiceAccount) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *ServiceAccountEventHandlerFuncs) UpdateServiceAccount(objOld, objNew *core_v1.ServiceAccount) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *ServiceAccountEventHandlerFuncs) GenericServiceAccount(obj *core_v1.ServiceAccount) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type ServiceAccountEventWatcher interface {
	AddEventHandler(ctx context.Context, h ServiceAccountEventHandler, predicates ...predicate.Predicate) error
}

type serviceAccountEventWatcher struct {
	watcher events.EventWatcher
}

func NewServiceAccountEventWatcher(name string, mgr manager.Manager) ServiceAccountEventWatcher {
	return &serviceAccountEventWatcher{
		watcher: events.NewWatcher(name, mgr, &core_v1.ServiceAccount{}),
	}
}

func (c *serviceAccountEventWatcher) AddEventHandler(ctx context.Context, h ServiceAccountEventHandler, predicates ...predicate.Predicate) error {
	handler := genericServiceAccountHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericServiceAccountHandler implements a generic events.EventHandler
type genericServiceAccountHandler struct {
	handler ServiceAccountEventHandler
}

func (h genericServiceAccountHandler) Create(object runtime.Object) error {
	obj, ok := object.(*core_v1.ServiceAccount)
	if !ok {
		return errors.Errorf("internal error: ServiceAccount handler received event for %T", object)
	}
	return h.handler.CreateServiceAccount(obj)
}

func (h genericServiceAccountHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.ServiceAccount)
	if !ok {
		return errors.Errorf("internal error: ServiceAccount handler received event for %T", object)
	}
	return h.handler.DeleteServiceAccount(obj)
}

func (h genericServiceAccountHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*core_v1.ServiceAccount)
	if !ok {
		return errors.Errorf("internal error: ServiceAccount handler received event for %T", old)
	}
	objNew, ok := new.(*core_v1.ServiceAccount)
	if !ok {
		return errors.Errorf("internal error: ServiceAccount handler received event for %T", new)
	}
	return h.handler.UpdateServiceAccount(objOld, objNew)
}

func (h genericServiceAccountHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.ServiceAccount)
	if !ok {
		return errors.Errorf("internal error: ServiceAccount handler received event for %T", object)
	}
	return h.handler.GenericServiceAccount(obj)
}

// Handle events for the ConfigMap Resource
type ConfigMapEventHandler interface {
	CreateConfigMap(obj *core_v1.ConfigMap) error
	UpdateConfigMap(old, new *core_v1.ConfigMap) error
	DeleteConfigMap(obj *core_v1.ConfigMap) error
	GenericConfigMap(obj *core_v1.ConfigMap) error
}

type ConfigMapEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.ConfigMap) error
	OnUpdate  func(old, new *core_v1.ConfigMap) error
	OnDelete  func(obj *core_v1.ConfigMap) error
	OnGeneric func(obj *core_v1.ConfigMap) error
}

func (f *ConfigMapEventHandlerFuncs) CreateConfigMap(obj *core_v1.ConfigMap) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *ConfigMapEventHandlerFuncs) DeleteConfigMap(obj *core_v1.ConfigMap) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *ConfigMapEventHandlerFuncs) UpdateConfigMap(objOld, objNew *core_v1.ConfigMap) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *ConfigMapEventHandlerFuncs) GenericConfigMap(obj *core_v1.ConfigMap) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type ConfigMapEventWatcher interface {
	AddEventHandler(ctx context.Context, h ConfigMapEventHandler, predicates ...predicate.Predicate) error
}

type configMapEventWatcher struct {
	watcher events.EventWatcher
}

func NewConfigMapEventWatcher(name string, mgr manager.Manager) ConfigMapEventWatcher {
	return &configMapEventWatcher{
		watcher: events.NewWatcher(name, mgr, &core_v1.ConfigMap{}),
	}
}

func (c *configMapEventWatcher) AddEventHandler(ctx context.Context, h ConfigMapEventHandler, predicates ...predicate.Predicate) error {
	handler := genericConfigMapHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericConfigMapHandler implements a generic events.EventHandler
type genericConfigMapHandler struct {
	handler ConfigMapEventHandler
}

func (h genericConfigMapHandler) Create(object runtime.Object) error {
	obj, ok := object.(*core_v1.ConfigMap)
	if !ok {
		return errors.Errorf("internal error: ConfigMap handler received event for %T", object)
	}
	return h.handler.CreateConfigMap(obj)
}

func (h genericConfigMapHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.ConfigMap)
	if !ok {
		return errors.Errorf("internal error: ConfigMap handler received event for %T", object)
	}
	return h.handler.DeleteConfigMap(obj)
}

func (h genericConfigMapHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*core_v1.ConfigMap)
	if !ok {
		return errors.Errorf("internal error: ConfigMap handler received event for %T", old)
	}
	objNew, ok := new.(*core_v1.ConfigMap)
	if !ok {
		return errors.Errorf("internal error: ConfigMap handler received event for %T", new)
	}
	return h.handler.UpdateConfigMap(objOld, objNew)
}

func (h genericConfigMapHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.ConfigMap)
	if !ok {
		return errors.Errorf("internal error: ConfigMap handler received event for %T", object)
	}
	return h.handler.GenericConfigMap(obj)
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

// Handle events for the Namespace Resource
type NamespaceEventHandler interface {
	CreateNamespace(obj *core_v1.Namespace) error
	UpdateNamespace(old, new *core_v1.Namespace) error
	DeleteNamespace(obj *core_v1.Namespace) error
	GenericNamespace(obj *core_v1.Namespace) error
}

type NamespaceEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Namespace) error
	OnUpdate  func(old, new *core_v1.Namespace) error
	OnDelete  func(obj *core_v1.Namespace) error
	OnGeneric func(obj *core_v1.Namespace) error
}

func (f *NamespaceEventHandlerFuncs) CreateNamespace(obj *core_v1.Namespace) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *NamespaceEventHandlerFuncs) DeleteNamespace(obj *core_v1.Namespace) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *NamespaceEventHandlerFuncs) UpdateNamespace(objOld, objNew *core_v1.Namespace) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *NamespaceEventHandlerFuncs) GenericNamespace(obj *core_v1.Namespace) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type NamespaceEventWatcher interface {
	AddEventHandler(ctx context.Context, h NamespaceEventHandler, predicates ...predicate.Predicate) error
}

type namespaceEventWatcher struct {
	watcher events.EventWatcher
}

func NewNamespaceEventWatcher(name string, mgr manager.Manager) NamespaceEventWatcher {
	return &namespaceEventWatcher{
		watcher: events.NewWatcher(name, mgr, &core_v1.Namespace{}),
	}
}

func (c *namespaceEventWatcher) AddEventHandler(ctx context.Context, h NamespaceEventHandler, predicates ...predicate.Predicate) error {
	handler := genericNamespaceHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericNamespaceHandler implements a generic events.EventHandler
type genericNamespaceHandler struct {
	handler NamespaceEventHandler
}

func (h genericNamespaceHandler) Create(object runtime.Object) error {
	obj, ok := object.(*core_v1.Namespace)
	if !ok {
		return errors.Errorf("internal error: Namespace handler received event for %T", object)
	}
	return h.handler.CreateNamespace(obj)
}

func (h genericNamespaceHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Namespace)
	if !ok {
		return errors.Errorf("internal error: Namespace handler received event for %T", object)
	}
	return h.handler.DeleteNamespace(obj)
}

func (h genericNamespaceHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*core_v1.Namespace)
	if !ok {
		return errors.Errorf("internal error: Namespace handler received event for %T", old)
	}
	objNew, ok := new.(*core_v1.Namespace)
	if !ok {
		return errors.Errorf("internal error: Namespace handler received event for %T", new)
	}
	return h.handler.UpdateNamespace(objOld, objNew)
}

func (h genericNamespaceHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Namespace)
	if !ok {
		return errors.Errorf("internal error: Namespace handler received event for %T", object)
	}
	return h.handler.GenericNamespace(obj)
}

// Handle events for the Node Resource
type NodeEventHandler interface {
	CreateNode(obj *core_v1.Node) error
	UpdateNode(old, new *core_v1.Node) error
	DeleteNode(obj *core_v1.Node) error
	GenericNode(obj *core_v1.Node) error
}

type NodeEventHandlerFuncs struct {
	OnCreate  func(obj *core_v1.Node) error
	OnUpdate  func(old, new *core_v1.Node) error
	OnDelete  func(obj *core_v1.Node) error
	OnGeneric func(obj *core_v1.Node) error
}

func (f *NodeEventHandlerFuncs) CreateNode(obj *core_v1.Node) error {
	if f.OnCreate == nil {
		return nil
	}
	return f.OnCreate(obj)
}

func (f *NodeEventHandlerFuncs) DeleteNode(obj *core_v1.Node) error {
	if f.OnDelete == nil {
		return nil
	}
	return f.OnDelete(obj)
}

func (f *NodeEventHandlerFuncs) UpdateNode(objOld, objNew *core_v1.Node) error {
	if f.OnUpdate == nil {
		return nil
	}
	return f.OnUpdate(objOld, objNew)
}

func (f *NodeEventHandlerFuncs) GenericNode(obj *core_v1.Node) error {
	if f.OnGeneric == nil {
		return nil
	}
	return f.OnGeneric(obj)
}

type NodeEventWatcher interface {
	AddEventHandler(ctx context.Context, h NodeEventHandler, predicates ...predicate.Predicate) error
}

type nodeEventWatcher struct {
	watcher events.EventWatcher
}

func NewNodeEventWatcher(name string, mgr manager.Manager) NodeEventWatcher {
	return &nodeEventWatcher{
		watcher: events.NewWatcher(name, mgr, &core_v1.Node{}),
	}
}

func (c *nodeEventWatcher) AddEventHandler(ctx context.Context, h NodeEventHandler, predicates ...predicate.Predicate) error {
	handler := genericNodeHandler{handler: h}
	if err := c.watcher.Watch(ctx, handler, predicates...); err != nil {
		return err
	}
	return nil
}

// genericNodeHandler implements a generic events.EventHandler
type genericNodeHandler struct {
	handler NodeEventHandler
}

func (h genericNodeHandler) Create(object runtime.Object) error {
	obj, ok := object.(*core_v1.Node)
	if !ok {
		return errors.Errorf("internal error: Node handler received event for %T", object)
	}
	return h.handler.CreateNode(obj)
}

func (h genericNodeHandler) Delete(object runtime.Object) error {
	obj, ok := object.(*core_v1.Node)
	if !ok {
		return errors.Errorf("internal error: Node handler received event for %T", object)
	}
	return h.handler.DeleteNode(obj)
}

func (h genericNodeHandler) Update(old, new runtime.Object) error {
	objOld, ok := old.(*core_v1.Node)
	if !ok {
		return errors.Errorf("internal error: Node handler received event for %T", old)
	}
	objNew, ok := new.(*core_v1.Node)
	if !ok {
		return errors.Errorf("internal error: Node handler received event for %T", new)
	}
	return h.handler.UpdateNode(objOld, objNew)
}

func (h genericNodeHandler) Generic(object runtime.Object) error {
	obj, ok := object.(*core_v1.Node)
	if !ok {
		return errors.Errorf("internal error: Node handler received event for %T", object)
	}
	return h.handler.GenericNode(obj)
}
