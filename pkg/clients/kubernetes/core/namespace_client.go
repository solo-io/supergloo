package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewGeneratedNamespaceClient(client kubernetes.Interface) NamespaceClient {
	return &generatedNamespaceClient{
		client: client,
	}
}

type generatedNamespaceClient struct {
	client kubernetes.Interface
}

func (g *generatedNamespaceClient) Get(ctx context.Context, name string) (*corev1.Namespace, error) {
	return g.client.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
}

func (g *generatedNamespaceClient) Delete(ctx context.Context, name string) error {
	return g.client.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{})
}

func (g *generatedNamespaceClient) Create(ctx context.Context, ns *corev1.Namespace) error {
	created, err := g.client.CoreV1().Namespaces().Create(ns)
	if err != nil {
		return err
	}
	*ns = *created
	return nil
}
