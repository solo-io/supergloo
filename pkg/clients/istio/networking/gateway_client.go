package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GatewayClientFactory func(client client.Client) GatewayClient

func NewGatewayClientFactory() GatewayClientFactory {
	return NewGatewayClient
}

func NewGatewayClient(client client.Client) GatewayClient {
	return &gatewayClient{
		client: client,
	}
}

type gatewayClient struct {
	client client.Client
}

func (g *gatewayClient) Create(ctx context.Context, gateway *v1alpha3.Gateway) error {
	return g.client.Create(ctx, gateway)
}

func (g *gatewayClient) Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha3.Gateway, error) {
	gateway := v1alpha3.Gateway{}
	err := g.client.Get(ctx, objKey, &gateway)
	if err != nil {
		return nil, err
	}

	return &gateway, nil
}

func (g *gatewayClient) Update(ctx context.Context, gateway *v1alpha3.Gateway) error {
	return g.client.Update(ctx, gateway)
}
