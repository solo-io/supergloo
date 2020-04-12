package kubernetes_apiext

import (
	"context"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CrdClientFactory func(cfg *rest.Config) (CustomResourceDefinitionClient, error)

func NewCrdClientFromConfigFactory() CrdClientFactory {
	return NewCrdClientFromConfig
}

func NewCrdClientFromConfig(cfg *rest.Config) (CustomResourceDefinitionClient, error) {
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	return NewCrdClient(dynamicClient), nil
}

func NewCrdClient(client client.Client) CustomResourceDefinitionClient {
	return &crdClient{client: client}
}

type crdClient struct {
	client client.Client
}

func (c *crdClient) Get(ctx context.Context, name string) (*v1beta1.CustomResourceDefinition, error) {
	csr := v1beta1.CustomResourceDefinition{}
	err := c.client.Get(ctx, client.ObjectKey{
		Name: name,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (c *crdClient) List(ctx context.Context) (*v1beta1.CustomResourceDefinitionList, error) {
	csr := v1beta1.CustomResourceDefinitionList{}
	err := c.client.List(ctx, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (c *crdClient) Delete(ctx context.Context, crd *v1beta1.CustomResourceDefinition) error {
	return c.client.Delete(ctx, crd)
}
