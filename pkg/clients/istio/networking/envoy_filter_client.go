package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EnvoyFilterClientFactory func(client client.Client) EnvoyFilterClient

func NewEnvoyFilterClientFactory() EnvoyFilterClientFactory {
	return NewEnvoyFilterClient
}

func NewEnvoyFilterClient(client client.Client) EnvoyFilterClient {
	return &envoyFilterClient{
		client: client,
	}
}

type envoyFilterClient struct {
	client client.Client
}

func (e *envoyFilterClient) Create(ctx context.Context, envoyFilter *v1alpha3.EnvoyFilter) error {
	return e.client.Create(ctx, envoyFilter)
}

func (e *envoyFilterClient) Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha3.EnvoyFilter, error) {
	filter := v1alpha3.EnvoyFilter{}
	err := e.client.Get(ctx, objKey, &filter)
	if err != nil {
		return nil, err
	}

	return &filter, nil
}

func (e *envoyFilterClient) UpsertSpec(ctx context.Context, envoyFilter *v1alpha3.EnvoyFilter) error {
	existing, err := e.Get(ctx, client.ObjectKey{Name: envoyFilter.Name, Namespace: envoyFilter.Namespace})
	if err != nil {
		if errors.IsNotFound(err) {
			return e.Create(ctx, envoyFilter)
		}
		return err
	}
	existing.Spec = envoyFilter.Spec
	return e.client.Update(ctx, existing)
}
