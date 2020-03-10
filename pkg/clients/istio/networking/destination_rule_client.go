package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DestinationRuleClientFactory func(client client.Client) DestinationRuleClient

func NewDestinationRuleClientFactory() DestinationRuleClientFactory {
	return NewDestinationRuleClient
}

func NewDestinationRuleClient(client client.Client) DestinationRuleClient {
	return &destinationRuleClient{
		client: client,
	}
}

type destinationRuleClient struct {
	client client.Client
}

func (d *destinationRuleClient) Create(ctx context.Context, destinationRule *v1alpha3.DestinationRule) error {
	return d.client.Create(ctx, destinationRule)
}

func (d *destinationRuleClient) Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha3.DestinationRule, error) {
	destinationRule := v1alpha3.DestinationRule{}
	err := d.client.Get(ctx, objKey, &destinationRule)
	if err != nil {
		return nil, err
	}

	return &destinationRule, nil
}
