// Definitions for the Kubernetes Controllers
package controller

import (
	"context"

	core_v1 "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconcile Upsert events for the Secret Resource.
// implemented by the user
type SecretReconciler interface {
	ReconcileSecret(obj *core_v1.Secret) (reconcile.Result, error)
}

// Reconcile deletion events for the Secret Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type SecretDeletionReconciler interface {
	ReconcileSecretDeletion(req reconcile.Request)
}

type SecretReconcilerFuncs struct {
	OnReconcileSecret         func(obj *core_v1.Secret) (reconcile.Result, error)
	OnReconcileSecretDeletion func(req reconcile.Request)
}

func (f *SecretReconcilerFuncs) ReconcileSecret(obj *core_v1.Secret) (reconcile.Result, error) {
	if f.OnReconcileSecret == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileSecret(obj)
}

func (f *SecretReconcilerFuncs) ReconcileSecretDeletion(req reconcile.Request) {
	if f.OnReconcileSecretDeletion == nil {
		return
	}
	f.OnReconcileSecretDeletion(req)
}

// Reconcile and finalize the Secret Resource
// implemented by the user
type SecretFinalizer interface {
	SecretReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	SecretFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeSecret(obj *core_v1.Secret) error
}

type SecretReconcileLoop interface {
	RunSecretReconciler(ctx context.Context, rec SecretReconciler, predicates ...predicate.Predicate) error
}

type secretReconcileLoop struct {
	loop reconcile.Loop
}

func NewSecretReconcileLoop(name string, mgr manager.Manager) SecretReconcileLoop {
	return &secretReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Secret{}),
	}
}

func (c *secretReconcileLoop) RunSecretReconciler(ctx context.Context, reconciler SecretReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericSecretReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(SecretFinalizer); ok {
		reconcilerWrapper = genericSecretFinalizer{
			genericSecretReconciler: genericReconciler,
			finalizingReconciler:    finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericSecretHandler implements a generic reconcile.Reconciler
type genericSecretReconciler struct {
	reconciler SecretReconciler
}

func (r genericSecretReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return r.reconciler.ReconcileSecret(obj)
}

func (r genericSecretReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(SecretDeletionReconciler); ok {
		deletionReconciler.ReconcileSecretDeletion(request)
	}
}

// genericSecretFinalizer implements a generic reconcile.FinalizingReconciler
type genericSecretFinalizer struct {
	genericSecretReconciler
	finalizingReconciler SecretFinalizer
}

func (r genericSecretFinalizer) FinalizerName() string {
	return r.finalizingReconciler.SecretFinalizerName()
}

func (r genericSecretFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Secret)
	if !ok {
		return errors.Errorf("internal error: Secret handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeSecret(obj)
}

// Reconcile Upsert events for the ServiceAccount Resource.
// implemented by the user
type ServiceAccountReconciler interface {
	ReconcileServiceAccount(obj *core_v1.ServiceAccount) (reconcile.Result, error)
}

// Reconcile deletion events for the ServiceAccount Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type ServiceAccountDeletionReconciler interface {
	ReconcileServiceAccountDeletion(req reconcile.Request)
}

type ServiceAccountReconcilerFuncs struct {
	OnReconcileServiceAccount         func(obj *core_v1.ServiceAccount) (reconcile.Result, error)
	OnReconcileServiceAccountDeletion func(req reconcile.Request)
}

func (f *ServiceAccountReconcilerFuncs) ReconcileServiceAccount(obj *core_v1.ServiceAccount) (reconcile.Result, error) {
	if f.OnReconcileServiceAccount == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileServiceAccount(obj)
}

func (f *ServiceAccountReconcilerFuncs) ReconcileServiceAccountDeletion(req reconcile.Request) {
	if f.OnReconcileServiceAccountDeletion == nil {
		return
	}
	f.OnReconcileServiceAccountDeletion(req)
}

// Reconcile and finalize the ServiceAccount Resource
// implemented by the user
type ServiceAccountFinalizer interface {
	ServiceAccountReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	ServiceAccountFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeServiceAccount(obj *core_v1.ServiceAccount) error
}

type ServiceAccountReconcileLoop interface {
	RunServiceAccountReconciler(ctx context.Context, rec ServiceAccountReconciler, predicates ...predicate.Predicate) error
}

type serviceAccountReconcileLoop struct {
	loop reconcile.Loop
}

func NewServiceAccountReconcileLoop(name string, mgr manager.Manager) ServiceAccountReconcileLoop {
	return &serviceAccountReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.ServiceAccount{}),
	}
}

func (c *serviceAccountReconcileLoop) RunServiceAccountReconciler(ctx context.Context, reconciler ServiceAccountReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericServiceAccountReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(ServiceAccountFinalizer); ok {
		reconcilerWrapper = genericServiceAccountFinalizer{
			genericServiceAccountReconciler: genericReconciler,
			finalizingReconciler:            finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericServiceAccountHandler implements a generic reconcile.Reconciler
type genericServiceAccountReconciler struct {
	reconciler ServiceAccountReconciler
}

func (r genericServiceAccountReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.ServiceAccount)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: ServiceAccount handler received event for %T", object)
	}
	return r.reconciler.ReconcileServiceAccount(obj)
}

func (r genericServiceAccountReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(ServiceAccountDeletionReconciler); ok {
		deletionReconciler.ReconcileServiceAccountDeletion(request)
	}
}

// genericServiceAccountFinalizer implements a generic reconcile.FinalizingReconciler
type genericServiceAccountFinalizer struct {
	genericServiceAccountReconciler
	finalizingReconciler ServiceAccountFinalizer
}

func (r genericServiceAccountFinalizer) FinalizerName() string {
	return r.finalizingReconciler.ServiceAccountFinalizerName()
}

func (r genericServiceAccountFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.ServiceAccount)
	if !ok {
		return errors.Errorf("internal error: ServiceAccount handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeServiceAccount(obj)
}

// Reconcile Upsert events for the ConfigMap Resource.
// implemented by the user
type ConfigMapReconciler interface {
	ReconcileConfigMap(obj *core_v1.ConfigMap) (reconcile.Result, error)
}

// Reconcile deletion events for the ConfigMap Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type ConfigMapDeletionReconciler interface {
	ReconcileConfigMapDeletion(req reconcile.Request)
}

type ConfigMapReconcilerFuncs struct {
	OnReconcileConfigMap         func(obj *core_v1.ConfigMap) (reconcile.Result, error)
	OnReconcileConfigMapDeletion func(req reconcile.Request)
}

func (f *ConfigMapReconcilerFuncs) ReconcileConfigMap(obj *core_v1.ConfigMap) (reconcile.Result, error) {
	if f.OnReconcileConfigMap == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileConfigMap(obj)
}

func (f *ConfigMapReconcilerFuncs) ReconcileConfigMapDeletion(req reconcile.Request) {
	if f.OnReconcileConfigMapDeletion == nil {
		return
	}
	f.OnReconcileConfigMapDeletion(req)
}

// Reconcile and finalize the ConfigMap Resource
// implemented by the user
type ConfigMapFinalizer interface {
	ConfigMapReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	ConfigMapFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeConfigMap(obj *core_v1.ConfigMap) error
}

type ConfigMapReconcileLoop interface {
	RunConfigMapReconciler(ctx context.Context, rec ConfigMapReconciler, predicates ...predicate.Predicate) error
}

type configMapReconcileLoop struct {
	loop reconcile.Loop
}

func NewConfigMapReconcileLoop(name string, mgr manager.Manager) ConfigMapReconcileLoop {
	return &configMapReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.ConfigMap{}),
	}
}

func (c *configMapReconcileLoop) RunConfigMapReconciler(ctx context.Context, reconciler ConfigMapReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericConfigMapReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(ConfigMapFinalizer); ok {
		reconcilerWrapper = genericConfigMapFinalizer{
			genericConfigMapReconciler: genericReconciler,
			finalizingReconciler:       finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericConfigMapHandler implements a generic reconcile.Reconciler
type genericConfigMapReconciler struct {
	reconciler ConfigMapReconciler
}

func (r genericConfigMapReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.ConfigMap)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: ConfigMap handler received event for %T", object)
	}
	return r.reconciler.ReconcileConfigMap(obj)
}

func (r genericConfigMapReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(ConfigMapDeletionReconciler); ok {
		deletionReconciler.ReconcileConfigMapDeletion(request)
	}
}

// genericConfigMapFinalizer implements a generic reconcile.FinalizingReconciler
type genericConfigMapFinalizer struct {
	genericConfigMapReconciler
	finalizingReconciler ConfigMapFinalizer
}

func (r genericConfigMapFinalizer) FinalizerName() string {
	return r.finalizingReconciler.ConfigMapFinalizerName()
}

func (r genericConfigMapFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.ConfigMap)
	if !ok {
		return errors.Errorf("internal error: ConfigMap handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeConfigMap(obj)
}

// Reconcile Upsert events for the Service Resource.
// implemented by the user
type ServiceReconciler interface {
	ReconcileService(obj *core_v1.Service) (reconcile.Result, error)
}

// Reconcile deletion events for the Service Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type ServiceDeletionReconciler interface {
	ReconcileServiceDeletion(req reconcile.Request)
}

type ServiceReconcilerFuncs struct {
	OnReconcileService         func(obj *core_v1.Service) (reconcile.Result, error)
	OnReconcileServiceDeletion func(req reconcile.Request)
}

func (f *ServiceReconcilerFuncs) ReconcileService(obj *core_v1.Service) (reconcile.Result, error) {
	if f.OnReconcileService == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileService(obj)
}

func (f *ServiceReconcilerFuncs) ReconcileServiceDeletion(req reconcile.Request) {
	if f.OnReconcileServiceDeletion == nil {
		return
	}
	f.OnReconcileServiceDeletion(req)
}

// Reconcile and finalize the Service Resource
// implemented by the user
type ServiceFinalizer interface {
	ServiceReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	ServiceFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeService(obj *core_v1.Service) error
}

type ServiceReconcileLoop interface {
	RunServiceReconciler(ctx context.Context, rec ServiceReconciler, predicates ...predicate.Predicate) error
}

type serviceReconcileLoop struct {
	loop reconcile.Loop
}

func NewServiceReconcileLoop(name string, mgr manager.Manager) ServiceReconcileLoop {
	return &serviceReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Service{}),
	}
}

func (c *serviceReconcileLoop) RunServiceReconciler(ctx context.Context, reconciler ServiceReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericServiceReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(ServiceFinalizer); ok {
		reconcilerWrapper = genericServiceFinalizer{
			genericServiceReconciler: genericReconciler,
			finalizingReconciler:     finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericServiceHandler implements a generic reconcile.Reconciler
type genericServiceReconciler struct {
	reconciler ServiceReconciler
}

func (r genericServiceReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return r.reconciler.ReconcileService(obj)
}

func (r genericServiceReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(ServiceDeletionReconciler); ok {
		deletionReconciler.ReconcileServiceDeletion(request)
	}
}

// genericServiceFinalizer implements a generic reconcile.FinalizingReconciler
type genericServiceFinalizer struct {
	genericServiceReconciler
	finalizingReconciler ServiceFinalizer
}

func (r genericServiceFinalizer) FinalizerName() string {
	return r.finalizingReconciler.ServiceFinalizerName()
}

func (r genericServiceFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Service)
	if !ok {
		return errors.Errorf("internal error: Service handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeService(obj)
}

// Reconcile Upsert events for the Pod Resource.
// implemented by the user
type PodReconciler interface {
	ReconcilePod(obj *core_v1.Pod) (reconcile.Result, error)
}

// Reconcile deletion events for the Pod Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type PodDeletionReconciler interface {
	ReconcilePodDeletion(req reconcile.Request)
}

type PodReconcilerFuncs struct {
	OnReconcilePod         func(obj *core_v1.Pod) (reconcile.Result, error)
	OnReconcilePodDeletion func(req reconcile.Request)
}

func (f *PodReconcilerFuncs) ReconcilePod(obj *core_v1.Pod) (reconcile.Result, error) {
	if f.OnReconcilePod == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcilePod(obj)
}

func (f *PodReconcilerFuncs) ReconcilePodDeletion(req reconcile.Request) {
	if f.OnReconcilePodDeletion == nil {
		return
	}
	f.OnReconcilePodDeletion(req)
}

// Reconcile and finalize the Pod Resource
// implemented by the user
type PodFinalizer interface {
	PodReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	PodFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizePod(obj *core_v1.Pod) error
}

type PodReconcileLoop interface {
	RunPodReconciler(ctx context.Context, rec PodReconciler, predicates ...predicate.Predicate) error
}

type podReconcileLoop struct {
	loop reconcile.Loop
}

func NewPodReconcileLoop(name string, mgr manager.Manager) PodReconcileLoop {
	return &podReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Pod{}),
	}
}

func (c *podReconcileLoop) RunPodReconciler(ctx context.Context, reconciler PodReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericPodReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(PodFinalizer); ok {
		reconcilerWrapper = genericPodFinalizer{
			genericPodReconciler: genericReconciler,
			finalizingReconciler: finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericPodHandler implements a generic reconcile.Reconciler
type genericPodReconciler struct {
	reconciler PodReconciler
}

func (r genericPodReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return r.reconciler.ReconcilePod(obj)
}

func (r genericPodReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(PodDeletionReconciler); ok {
		deletionReconciler.ReconcilePodDeletion(request)
	}
}

// genericPodFinalizer implements a generic reconcile.FinalizingReconciler
type genericPodFinalizer struct {
	genericPodReconciler
	finalizingReconciler PodFinalizer
}

func (r genericPodFinalizer) FinalizerName() string {
	return r.finalizingReconciler.PodFinalizerName()
}

func (r genericPodFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Pod)
	if !ok {
		return errors.Errorf("internal error: Pod handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizePod(obj)
}

// Reconcile Upsert events for the Namespace Resource.
// implemented by the user
type NamespaceReconciler interface {
	ReconcileNamespace(obj *core_v1.Namespace) (reconcile.Result, error)
}

// Reconcile deletion events for the Namespace Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type NamespaceDeletionReconciler interface {
	ReconcileNamespaceDeletion(req reconcile.Request)
}

type NamespaceReconcilerFuncs struct {
	OnReconcileNamespace         func(obj *core_v1.Namespace) (reconcile.Result, error)
	OnReconcileNamespaceDeletion func(req reconcile.Request)
}

func (f *NamespaceReconcilerFuncs) ReconcileNamespace(obj *core_v1.Namespace) (reconcile.Result, error) {
	if f.OnReconcileNamespace == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileNamespace(obj)
}

func (f *NamespaceReconcilerFuncs) ReconcileNamespaceDeletion(req reconcile.Request) {
	if f.OnReconcileNamespaceDeletion == nil {
		return
	}
	f.OnReconcileNamespaceDeletion(req)
}

// Reconcile and finalize the Namespace Resource
// implemented by the user
type NamespaceFinalizer interface {
	NamespaceReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	NamespaceFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeNamespace(obj *core_v1.Namespace) error
}

type NamespaceReconcileLoop interface {
	RunNamespaceReconciler(ctx context.Context, rec NamespaceReconciler, predicates ...predicate.Predicate) error
}

type namespaceReconcileLoop struct {
	loop reconcile.Loop
}

func NewNamespaceReconcileLoop(name string, mgr manager.Manager) NamespaceReconcileLoop {
	return &namespaceReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Namespace{}),
	}
}

func (c *namespaceReconcileLoop) RunNamespaceReconciler(ctx context.Context, reconciler NamespaceReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericNamespaceReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(NamespaceFinalizer); ok {
		reconcilerWrapper = genericNamespaceFinalizer{
			genericNamespaceReconciler: genericReconciler,
			finalizingReconciler:       finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericNamespaceHandler implements a generic reconcile.Reconciler
type genericNamespaceReconciler struct {
	reconciler NamespaceReconciler
}

func (r genericNamespaceReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Namespace)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Namespace handler received event for %T", object)
	}
	return r.reconciler.ReconcileNamespace(obj)
}

func (r genericNamespaceReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(NamespaceDeletionReconciler); ok {
		deletionReconciler.ReconcileNamespaceDeletion(request)
	}
}

// genericNamespaceFinalizer implements a generic reconcile.FinalizingReconciler
type genericNamespaceFinalizer struct {
	genericNamespaceReconciler
	finalizingReconciler NamespaceFinalizer
}

func (r genericNamespaceFinalizer) FinalizerName() string {
	return r.finalizingReconciler.NamespaceFinalizerName()
}

func (r genericNamespaceFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Namespace)
	if !ok {
		return errors.Errorf("internal error: Namespace handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeNamespace(obj)
}

// Reconcile Upsert events for the Node Resource.
// implemented by the user
type NodeReconciler interface {
	ReconcileNode(obj *core_v1.Node) (reconcile.Result, error)
}

// Reconcile deletion events for the Node Resource.
// Deletion receives a reconcile.Request as we cannot guarantee the last state of the object
// before being deleted.
// implemented by the user
type NodeDeletionReconciler interface {
	ReconcileNodeDeletion(req reconcile.Request)
}

type NodeReconcilerFuncs struct {
	OnReconcileNode         func(obj *core_v1.Node) (reconcile.Result, error)
	OnReconcileNodeDeletion func(req reconcile.Request)
}

func (f *NodeReconcilerFuncs) ReconcileNode(obj *core_v1.Node) (reconcile.Result, error) {
	if f.OnReconcileNode == nil {
		return reconcile.Result{}, nil
	}
	return f.OnReconcileNode(obj)
}

func (f *NodeReconcilerFuncs) ReconcileNodeDeletion(req reconcile.Request) {
	if f.OnReconcileNodeDeletion == nil {
		return
	}
	f.OnReconcileNodeDeletion(req)
}

// Reconcile and finalize the Node Resource
// implemented by the user
type NodeFinalizer interface {
	NodeReconciler

	// name of the finalizer used by this handler.
	// finalizer names should be unique for a single task
	NodeFinalizerName() string

	// finalize the object before it is deleted.
	// Watchers created with a finalizing handler will a
	FinalizeNode(obj *core_v1.Node) error
}

type NodeReconcileLoop interface {
	RunNodeReconciler(ctx context.Context, rec NodeReconciler, predicates ...predicate.Predicate) error
}

type nodeReconcileLoop struct {
	loop reconcile.Loop
}

func NewNodeReconcileLoop(name string, mgr manager.Manager) NodeReconcileLoop {
	return &nodeReconcileLoop{
		loop: reconcile.NewLoop(name, mgr, &core_v1.Node{}),
	}
}

func (c *nodeReconcileLoop) RunNodeReconciler(ctx context.Context, reconciler NodeReconciler, predicates ...predicate.Predicate) error {
	genericReconciler := genericNodeReconciler{
		reconciler: reconciler,
	}

	var reconcilerWrapper reconcile.Reconciler
	if finalizingReconciler, ok := reconciler.(NodeFinalizer); ok {
		reconcilerWrapper = genericNodeFinalizer{
			genericNodeReconciler: genericReconciler,
			finalizingReconciler:  finalizingReconciler,
		}
	} else {
		reconcilerWrapper = genericReconciler
	}
	return c.loop.RunReconciler(ctx, reconcilerWrapper, predicates...)
}

// genericNodeHandler implements a generic reconcile.Reconciler
type genericNodeReconciler struct {
	reconciler NodeReconciler
}

func (r genericNodeReconciler) Reconcile(object ezkube.Object) (reconcile.Result, error) {
	obj, ok := object.(*core_v1.Node)
	if !ok {
		return reconcile.Result{}, errors.Errorf("internal error: Node handler received event for %T", object)
	}
	return r.reconciler.ReconcileNode(obj)
}

func (r genericNodeReconciler) ReconcileDeletion(request reconcile.Request) {
	if deletionReconciler, ok := r.reconciler.(NodeDeletionReconciler); ok {
		deletionReconciler.ReconcileNodeDeletion(request)
	}
}

// genericNodeFinalizer implements a generic reconcile.FinalizingReconciler
type genericNodeFinalizer struct {
	genericNodeReconciler
	finalizingReconciler NodeFinalizer
}

func (r genericNodeFinalizer) FinalizerName() string {
	return r.finalizingReconciler.NodeFinalizerName()
}

func (r genericNodeFinalizer) Finalize(object ezkube.Object) error {
	obj, ok := object.(*core_v1.Node)
	if !ok {
		return errors.Errorf("internal error: Node handler received event for %T", object)
	}
	return r.finalizingReconciler.FinalizeNode(obj)
}
