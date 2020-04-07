package zephyr_networking

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type trafficPolicyClient struct {
	client client.Client
}

func NewTrafficPolicyClient(client client.Client) TrafficPolicyClient {
	return &trafficPolicyClient{client: client}
}

func (d *trafficPolicyClient) Get(ctx context.Context, name string, namespace string) (*v1alpha1.TrafficPolicy, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	trafficPolicy := &v1alpha1.TrafficPolicy{}
	err := d.client.Get(ctx, key, trafficPolicy)
	if err != nil {
		return nil, err
	}
	return trafficPolicy, nil
}

func (d *trafficPolicyClient) Create(ctx context.Context, trafficPolicy *v1alpha1.TrafficPolicy, options ...client.CreateOption) error {
	return d.client.Create(ctx, trafficPolicy, options...)
}

func (d *trafficPolicyClient) Update(ctx context.Context, trafficPolicy *v1alpha1.TrafficPolicy, options ...client.UpdateOption) error {
	return d.client.Update(ctx, trafficPolicy, options...)
}

func (d *trafficPolicyClient) UpdateStatus(ctx context.Context, trafficPolicy *v1alpha1.TrafficPolicy, options ...client.UpdateOption) error {
	return d.client.Status().Update(ctx, trafficPolicy, options...)
}

func (d *trafficPolicyClient) List(ctx context.Context, options ...client.ListOption) (*v1alpha1.TrafficPolicyList, error) {
	list := v1alpha1.TrafficPolicyList{}
	err := d.client.List(ctx, &list, options...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
