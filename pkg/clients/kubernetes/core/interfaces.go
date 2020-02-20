package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type ServiceClient interface {
	// include client.InNamespace(ns) in the options varargs list to specify a namespace
	// always returns a non-nil ServiceList, but its Items field may be empty
	List(ctx context.Context, options ...client.ListOption) (*corev1.ServiceList, error)
}

type SecretsClient interface {
	Update(ctx context.Context, csr *corev1.Secret) error
	Get(ctx context.Context, name, namespace string) (*corev1.Secret, error)
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.SecretList, error)
}
