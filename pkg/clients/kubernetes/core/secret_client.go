package kubernetes_core

import (
	"context"

	v1 "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewExtendedSecretClient(client client.Client) ExtendedSecretClient {
	return &extendedSecretClient{secretClient: v1.NewSecretClient(client)}
}

type extendedSecretClient struct {
	secretClient v1.SecretClient
}

func (e *extendedSecretClient) CreateSecret(ctx context.Context, secret *corev1.Secret, opts ...client.CreateOption) error {
	return e.secretClient.CreateSecret(ctx, secret, opts...)
}

func (e *extendedSecretClient) UpdateSecret(ctx context.Context, secret *corev1.Secret, opts ...client.UpdateOption) error {
	return e.secretClient.UpdateSecret(ctx, secret, opts...)
}

func (e *extendedSecretClient) UpsertData(ctx context.Context, secret *corev1.Secret) error {
	existing, err := e.secretClient.GetSecret(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace})
	if err != nil {
		if errors.IsNotFound(err) {
			return e.secretClient.CreateSecret(ctx, secret)
		}
		return err
	}
	existing.Data = secret.Data
	existing.StringData = secret.StringData
	return e.secretClient.UpdateSecret(ctx, existing)
}

func (e *extendedSecretClient) GetSecret(ctx context.Context, key client.ObjectKey) (*corev1.Secret, error) {
	return e.secretClient.GetSecret(ctx, key)
}

func (e *extendedSecretClient) ListSecret(ctx context.Context, opts ...client.ListOption) (*corev1.SecretList, error) {
	return e.secretClient.ListSecret(ctx, opts...)
}

func (e *extendedSecretClient) DeleteSecret(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error {
	return e.secretClient.DeleteSecret(ctx, key, opts...)
}

func (e *extendedSecretClient) PatchSecret(ctx context.Context, obj *corev1.Secret, patch client.Patch, opts ...client.PatchOption) error {
	return e.secretClient.PatchSecret(ctx, obj, patch, opts...)
}

func (e *extendedSecretClient) DeleteAllOfSecret(ctx context.Context, opts ...client.DeleteAllOfOption) error {
	return e.secretClient.DeleteAllOfSecret(ctx, opts...)
}

func (e *extendedSecretClient) UpdateSecretStatus(ctx context.Context, obj *corev1.Secret, opts ...client.UpdateOption) error {
	return e.secretClient.UpdateSecretStatus(ctx, obj, opts...)
}

func (e *extendedSecretClient) PatchSecretStatus(ctx context.Context, obj *corev1.Secret, patch client.Patch, opts ...client.PatchOption) error {
	return e.secretClient.PatchSecretStatus(ctx, obj, patch, opts...)
}
