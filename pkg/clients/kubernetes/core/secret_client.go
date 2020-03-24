package kubernetes_core

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSecretsClient(dynamicClient client.Client) SecretsClient {
	return &secretsClient{dynamicClient: dynamicClient}
}

type secretsClient struct {
	cluster       string
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

type GeneratedSecretClientFactory func(cfg *rest.Config) (SecretsClient, error)

func GeneratedSecretClientFactoryProvider() GeneratedSecretClientFactory {
	return func(cfg *rest.Config) (SecretsClient, error) {
		cs, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}

		return NewGeneratedSecretsClient(cs), nil
	}
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

func (s *secretsGeneratedClient) UpsertData(_ context.Context, secret *corev1.Secret) error {
	upserted, err := s.client.Secrets(secret.GetNamespace()).Create(secret)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			upserted, err = s.client.Secrets(secret.GetNamespace()).Update(secret)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	*secret = *upserted
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

func (s *secretsGeneratedClient) List(ctx context.Context, namespace string, labels map[string]string) (*corev1.SecretList, error) {
	var labelPairs []string
	for k, v := range labels {
		labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", k, v))
	}
	listOptions := metav1.ListOptions{
		LabelSelector: strings.Join(labelPairs, ","),
	}

	return s.client.Secrets(namespace).List(listOptions)
}

func (s *secretsGeneratedClient) Delete(ctx context.Context, secret *corev1.Secret) error {
	return s.client.Secrets(secret.GetNamespace()).Delete(secret.GetName(), &metav1.DeleteOptions{})
}
