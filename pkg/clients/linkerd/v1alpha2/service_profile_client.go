package v1alpha2

import (
	"context"

	v1alpha2 "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type serviceProfileClient struct {
	client client.Client
}

type ServiceProfileClientFactory func(client client.Client) ServiceProfileClient

func ServiceProfileClientFactoryProvider() ServiceProfileClientFactory {
	return NewServiceProfileClient
}

func NewServiceProfileClient(client client.Client) ServiceProfileClient {
	return &serviceProfileClient{client: client}
}

func (a *serviceProfileClient) Get(ctx context.Context, key client.ObjectKey) (*v1alpha2.ServiceProfile, error) {
	serviceProfile := v1alpha2.ServiceProfile{}
	err := a.client.Get(ctx, key, &serviceProfile)
	if err != nil {
		return nil, err
	}
	return &serviceProfile, nil
}

func (a *serviceProfileClient) List(ctx context.Context, options ...client.ListOption) (*v1alpha2.ServiceProfileList, error) {
	serviceProfileList := v1alpha2.ServiceProfileList{}
	err := a.client.List(ctx, &serviceProfileList, options...)
	if err != nil {
		return nil, err
	}
	return &serviceProfileList, nil
}

func (a *serviceProfileClient) UpsertSpec(ctx context.Context, serviceProfile *v1alpha2.ServiceProfile) error {
	key := client.ObjectKey{Name: serviceProfile.GetName(), Namespace: serviceProfile.GetNamespace()}
	existingAuthPolicy, err := a.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return a.Create(ctx, serviceProfile)
		}
		return err
	}
	existingAuthPolicy.Spec = serviceProfile.Spec
	return a.Update(ctx, existingAuthPolicy)
}

func (a *serviceProfileClient) Create(ctx context.Context, serviceProfile *v1alpha2.ServiceProfile, options ...client.CreateOption) error {
	return a.client.Create(ctx, serviceProfile, options...)
}

func (a *serviceProfileClient) Update(ctx context.Context, serviceProfile *v1alpha2.ServiceProfile, options ...client.UpdateOption) error {
	return a.client.Update(ctx, serviceProfile, options...)
}

func (a *serviceProfileClient) Delete(ctx context.Context, key client.ObjectKey) error {
	authPolicy, err := a.Get(ctx, key)
	if err != nil {
		return err
	}
	return a.client.Delete(ctx, authPolicy)
}
