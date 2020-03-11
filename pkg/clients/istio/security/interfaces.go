package security

import (
	"context"

	"istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AuthorizationPolicyClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1beta1.AuthorizationPolicy, error)
	Create(ctx context.Context, virtualService *v1beta1.AuthorizationPolicy, options ...client.CreateOption) error
	Update(ctx context.Context, virtualService *v1beta1.AuthorizationPolicy, options ...client.UpdateOption) error
	// Create the AuthorizationPolicy if it does not exist, otherwise update
	Upsert(ctx context.Context, virtualService *v1beta1.AuthorizationPolicy) error
}
