package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewNamespaceClientForConfig(cfg *rest.Config) (NamespaceClient, error) {
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &namespaceClient{
		client: dynamicClient,
	}, nil
}

func NewNamespaceClient(dynamicClient client.Client) NamespaceClient {
	return &namespaceClient{
		client: dynamicClient,
	}
}

type namespaceClient struct {
	client client.Client
}

func (g *namespaceClient) Get(ctx context.Context, name string) (*corev1.Namespace, error) {
	ns := corev1.Namespace{}
	err := g.client.Get(ctx, client.ObjectKey{Name: name}, &ns)
	return &ns, err
}

func (g *namespaceClient) Delete(ctx context.Context, ns *corev1.Namespace) error {
	return g.client.Delete(ctx, ns)
}

func (g *namespaceClient) Create(ctx context.Context, ns *corev1.Namespace) error {
	return g.client.Create(ctx, ns)
}
