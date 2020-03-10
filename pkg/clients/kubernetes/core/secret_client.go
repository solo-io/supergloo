package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
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

func NewGeneratedSecretsClient(client kubernetes.Interface) SecretsClient {
	return &secretsGeneratedClient{client: client.CoreV1()}
}

type secretsGeneratedClient struct {
	client k8sclientv1.SecretsGetter
}

func (s *secretsGeneratedClient) Update(_ context.Context, secret *corev1.Secret, opts ...client.UpdateOption) error {
	updated, err := s.client.Secrets(secret.GetNamespace()).Update(secret)
	if err != nil {
		return err
	}
	*secret = *updated
	return nil
}

func (s *secretsGeneratedClient) Create(_ context.Context, secret *corev1.Secret, opts ...client.CreateOption) error {
	updated, err := s.client.Secrets(secret.GetNamespace()).Create(secret)
	if err != nil {
		return err
	}
	*secret = *updated
	return nil
}

func (s *secretsGeneratedClient) Get(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	return s.client.Secrets(namespace).Get(name, metav1.GetOptions{})
}

func (s *secretsGeneratedClient) List(ctx context.Context, opts metav1.ListOptions) (*corev1.SecretList, error) {
	return s.client.Secrets("").List(metav1.ListOptions{})
}
