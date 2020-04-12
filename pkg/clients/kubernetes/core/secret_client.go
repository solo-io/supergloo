package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSecretsClient(dynamicClient client.Client) SecretClient {
	return &secretsClient{dynamicClient: dynamicClient}
}

type secretsClient struct {
	dynamicClient client.Client
}

func (s *secretsClient) Create(ctx context.Context, secret *corev1.Secret, opts ...client.CreateOption) error {
	return s.dynamicClient.Create(ctx, secret, opts...)
}

func (s *secretsClient) Update(ctx context.Context, secret *corev1.Secret, opts ...client.UpdateOption) error {
	return s.dynamicClient.Update(ctx, secret, opts...)
}

func (s *secretsClient) UpsertData(ctx context.Context, secret *corev1.Secret) error {
	existing, err := s.Get(ctx, secret.Name, secret.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return s.dynamicClient.Create(ctx, secret)
		}
		return err
	}
	existing.Data = secret.Data
	existing.StringData = secret.StringData
	return s.dynamicClient.Update(ctx, existing)
}

func (s *secretsClient) Get(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	err := s.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &secret)
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

func (c *secretsClient) List(ctx context.Context, namespace string, labels map[string]string) (*corev1.SecretList, error) {
	list := corev1.SecretList{}
	var opts = []client.ListOption{client.InNamespace(namespace), client.MatchingLabels(labels)}
	err := c.dynamicClient.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *secretsClient) Delete(ctx context.Context, secret *corev1.Secret) error {
	return c.dynamicClient.Delete(ctx, secret)
}

func NewSecretsClientForConfig(cfg *rest.Config) (SecretClient, error) {
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &secretsClient{dynamicClient: dynamicClient}, nil
}
