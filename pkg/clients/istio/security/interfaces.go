package security

import (
	"context"

	"istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mock/mock_interfaces.go

type AuthorizationPolicyClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1beta1.AuthorizationPolicy, error)
	Create(ctx context.Context, authPolicy *v1beta1.AuthorizationPolicy, options ...client.CreateOption) error
	Update(ctx context.Context, authPolicy *v1beta1.AuthorizationPolicy, options ...client.UpdateOption) error
	// Create the AuthorizationPolicy if it does not exist, otherwise update
	UpsertSpec(ctx context.Context, authPolicy *v1beta1.AuthorizationPolicy) error
}
