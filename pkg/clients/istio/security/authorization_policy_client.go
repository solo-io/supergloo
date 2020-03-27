package security

import (
	"context"

	"istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type authorizationPolicyClient struct {
	client client.Client
}

type AuthorizationPolicyClientFactory func(client client.Client) AuthorizationPolicyClient

func AuthorizationPolicyClientFactoryProvider() AuthorizationPolicyClientFactory {
	return NewAuthorizationPolicyClient
}

func NewAuthorizationPolicyClient(client client.Client) AuthorizationPolicyClient {
	return &authorizationPolicyClient{client: client}
}

func (a *authorizationPolicyClient) Get(ctx context.Context, key client.ObjectKey) (*v1beta1.AuthorizationPolicy, error) {
	authorizationPolicy := v1beta1.AuthorizationPolicy{}
	err := a.client.Get(ctx, key, &authorizationPolicy)
	if err != nil {
		return nil, err
	}
	return &authorizationPolicy, nil
}

func (a *authorizationPolicyClient) UpsertSpec(ctx context.Context, authorizationPolicy *v1beta1.AuthorizationPolicy) error {
	key := client.ObjectKey{Name: authorizationPolicy.GetName(), Namespace: authorizationPolicy.GetNamespace()}
	existingAuthPolicy, err := a.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return a.Create(ctx, authorizationPolicy)
		}
		return err
	}
	existingAuthPolicy.Spec = authorizationPolicy.Spec
	return a.Update(ctx, existingAuthPolicy)
}

func (a *authorizationPolicyClient) Create(ctx context.Context, authorizationPolicy *v1beta1.AuthorizationPolicy, options ...client.CreateOption) error {
	return a.client.Create(ctx, authorizationPolicy, options...)
}

func (a *authorizationPolicyClient) Update(ctx context.Context, authorizationPolicy *v1beta1.AuthorizationPolicy, options ...client.UpdateOption) error {
	return a.client.Update(ctx, authorizationPolicy, options...)
}
