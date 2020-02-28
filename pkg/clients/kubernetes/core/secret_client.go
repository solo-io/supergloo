package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSecretsClient(dynamicClient client.Client) SecretsClient {
	return &secretsClient{dynamicClient: dynamicClient}
}

type secretsClient struct {
	cluster       string
	dynamicClient client.Client
}

func (c *secretsClient) Create(ctx context.Context, secret *corev1.Secret, opts ...client.CreateOption) error {
	return c.dynamicClient.Create(ctx, secret, opts...)
}

func (c *secretsClient) Update(ctx context.Context, secret *corev1.Secret, opts ...client.UpdateOption) error {
	return c.dynamicClient.Update(ctx, secret, opts...)
}

func (c *secretsClient) Get(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	csr := corev1.Secret{}
	err := c.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (c *secretsClient) List(ctx context.Context, opts metav1.ListOptions) (*corev1.SecretList, error) {
	list := corev1.SecretList{}
	err := c.dynamicClient.List(ctx, &list, &client.ListOptions{Raw: &opts})
	if err != nil {
		return nil, err
	}
	return &list, nil
}
