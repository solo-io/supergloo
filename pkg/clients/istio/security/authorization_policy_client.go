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

func (v *authorizationPolicyClient) Get(ctx context.Context, key client.ObjectKey) (*v1beta1.AuthorizationPolicy, error) {
	authorizationPolicy := v1beta1.AuthorizationPolicy{}
	err := v.client.Get(ctx, key, &authorizationPolicy)
	if err != nil {
		return nil, err
	}
	return &authorizationPolicy, nil
}

func (v *authorizationPolicyClient) Upsert(ctx context.Context, authorizationPolicy *v1beta1.AuthorizationPolicy) error {
	key := client.ObjectKey{Name: authorizationPolicy.GetName(), Namespace: authorizationPolicy.GetNamespace()}
	_, err := v.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return v.Create(ctx, authorizationPolicy)
		}
		return err
	}
	return v.Update(ctx, authorizationPolicy)
}

func (v *authorizationPolicyClient) Create(ctx context.Context, authorizationPolicy *v1beta1.AuthorizationPolicy, options ...client.CreateOption) error {
	return v.client.Create(ctx, authorizationPolicy, options...)
}

func (v *authorizationPolicyClient) Update(ctx context.Context, authorizationPolicy *v1beta1.AuthorizationPolicy, options ...client.UpdateOption) error {
	return v.client.Update(ctx, authorizationPolicy, options...)
}
