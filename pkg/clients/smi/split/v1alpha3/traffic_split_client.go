package v1alpha3

import (
	"context"

	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha3"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type trafficSplitClient struct {
	client client.Client
}

type TrafficSplitClientFactory func(client client.Client) TrafficSplitClient

func TrafficSplitClientFactoryProvider() TrafficSplitClientFactory {
	return NewTrafficSplitClient
}

func NewTrafficSplitClient(client client.Client) TrafficSplitClient {
	return &trafficSplitClient{client: client}
}

func (a *trafficSplitClient) Get(ctx context.Context, key client.ObjectKey) (*v1alpha3.TrafficSplit, error) {
	trafficSplit := v1alpha3.TrafficSplit{}
	err := a.client.Get(ctx, key, &trafficSplit)
	if err != nil {
		return nil, err
	}
	return &trafficSplit, nil
}

func (a *trafficSplitClient) UpsertSpec(ctx context.Context, trafficSplit *v1alpha3.TrafficSplit) error {
	key := client.ObjectKey{Name: trafficSplit.GetName(), Namespace: trafficSplit.GetNamespace()}
	existingAuthPolicy, err := a.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return a.Create(ctx, trafficSplit)
		}
		return err
	}
	existingAuthPolicy.Spec = trafficSplit.Spec
	return a.Update(ctx, existingAuthPolicy)
}

func (a *trafficSplitClient) Create(ctx context.Context, trafficSplit *v1alpha3.TrafficSplit, options ...client.CreateOption) error {
	return a.client.Create(ctx, trafficSplit, options...)
}

func (a *trafficSplitClient) Update(ctx context.Context, trafficSplit *v1alpha3.TrafficSplit, options ...client.UpdateOption) error {
	return a.client.Update(ctx, trafficSplit, options...)
}

func (a *trafficSplitClient) Delete(ctx context.Context, key client.ObjectKey) error {
	authPolicy, err := a.Get(ctx, key)
	if err != nil {
		return err
	}
	return a.client.Delete(ctx, authPolicy)
}
