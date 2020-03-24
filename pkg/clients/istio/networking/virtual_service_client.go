package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type virtualServiceClient struct {
	client client.Client
}

type VirtualServiceClientFactory func(client client.Client) VirtualServiceClient

func VirtualServiceClientFactoryProvider() VirtualServiceClientFactory {
	return NewVirtualServiceClient
}

func NewVirtualServiceClient(client client.Client) VirtualServiceClient {
	return &virtualServiceClient{client: client}
}

func (v *virtualServiceClient) Get(ctx context.Context, key client.ObjectKey) (*v1alpha3.VirtualService, error) {
	virtualService := v1alpha3.VirtualService{}
	err := v.client.Get(ctx, key, &virtualService)
	if err != nil {
		return nil, err
	}
	return &virtualService, nil
}

func (v *virtualServiceClient) UpsertSpec(ctx context.Context, virtualService *v1alpha3.VirtualService) error {
	key := client.ObjectKey{Name: virtualService.GetName(), Namespace: virtualService.GetNamespace()}
	existing, err := v.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return v.Create(ctx, virtualService)
		}
		return err
	}
	existing.Spec = virtualService.Spec
	return v.Update(ctx, existing)
}

func (v *virtualServiceClient) Create(ctx context.Context, virtualService *v1alpha3.VirtualService, options ...client.CreateOption) error {
	return v.client.Create(ctx, virtualService, options...)
}

func (v *virtualServiceClient) Update(ctx context.Context, virtualService *v1alpha3.VirtualService, options ...client.UpdateOption) error {
	return v.client.Update(ctx, virtualService, options...)
}
