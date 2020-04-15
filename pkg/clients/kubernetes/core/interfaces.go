package kubernetes_core

import (
	"context"

	v1 "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type ExtendedSecretClient interface {
	v1.SecretClient
	UpsertData(ctx context.Context, secret *corev1.Secret) error
}
