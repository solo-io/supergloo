package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DestinationRuleClientFactoryProvider() DestinationRuleClientFactory {
	return NewDestinationRuleClient
}

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

func (d *destinationRuleClient) Update(ctx context.Context, destinationRule *v1alpha3.DestinationRule, options ...client.UpdateOption) error {
	return d.client.Update(ctx, destinationRule, options...)
}

func (d *destinationRuleClient) Upsert(ctx context.Context, destinationRule *v1alpha3.DestinationRule) error {
	key := client.ObjectKey{Name: destinationRule.GetName(), Namespace: destinationRule.GetNamespace()}
	_, err := d.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return d.Create(ctx, destinationRule)
		}
		return err
	}
	return d.Update(ctx, destinationRule)
}
