package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type ServiceClient interface {
	Get(ctx context.Context, name, namespace string) (*corev1.Service, error)
	// include client.InNamespace(ns) in the options varargs list to specify a namespace
	// always returns a non-nil ServiceList, but its Items field may be empty
	List(ctx context.Context, options ...client.ListOption) (*corev1.ServiceList, error)
}

type PodClient interface {
	Get(ctx context.Context, name, namespace string) (*corev1.Pod, error)
	List(ctx context.Context, options ...client.ListOption) (*corev1.PodList, error)
}

type NodeClient interface {
	Get(ctx context.Context, name string) (*corev1.Node, error)
	List(ctx context.Context, options ...client.ListOption) (*corev1.NodeList, error)
}

type SecretsClient interface {
	Create(ctx context.Context, secret *corev1.Secret, opts ...client.CreateOption) error
	Update(ctx context.Context, secret *corev1.Secret, opts ...client.UpdateOption) error
	UpsertData(ctx context.Context, secret *corev1.Secret) error
	Get(ctx context.Context, name, namespace string) (*corev1.Secret, error)
	List(ctx context.Context, namespace string, labels map[string]string) (*corev1.SecretList, error)
	Delete(ctx context.Context, secret *corev1.Secret) error
}

type ServiceAccountClient interface {
	// create the service account in the namespace on the resource's ObjectMeta
	Create(ctx context.Context, serviceAccount *corev1.ServiceAccount) error

	Get(ctx context.Context, name, namespace string) (*corev1.ServiceAccount, error)

	// update the service account in the namespace on the resource's ObjectMeta
	Update(ctx context.Context, serviceAccount *corev1.ServiceAccount) error
}

type ConfigMapClient interface {
	Create(ctx context.Context, configMap *corev1.ConfigMap) error
	Get(ctx context.Context, objKey client.ObjectKey) (*corev1.ConfigMap, error)
	Update(ctx context.Context, configMap *corev1.ConfigMap) error
}

type NamespaceClient interface {
	Create(ctx context.Context, ns *corev1.Namespace) error
	Get(ctx context.Context, name string) (*corev1.Namespace, error)
	Delete(ctx context.Context, name string) error
}
