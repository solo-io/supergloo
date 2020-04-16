package zephyr_networking

import (
	"context"

	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type trafficPolicyClient struct {
	client client.Client
}

func NewTrafficPolicyClientForConfig(cfg *rest.Config) (TrafficPolicyClient, error) {
	if err := networking_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &trafficPolicyClient{client: dynamicClient}, nil
}

func NewTrafficPolicyClient(client client.Client) TrafficPolicyClient {
	return &trafficPolicyClient{client: client}
}

func (d *trafficPolicyClient) Get(ctx context.Context, name string, namespace string) (*networking_v1alpha1.TrafficPolicy, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	trafficPolicy := &networking_v1alpha1.TrafficPolicy{}
	err := d.client.Get(ctx, key, trafficPolicy)
	if err != nil {
		return nil, err
	}
	return trafficPolicy, nil
}

func (d *trafficPolicyClient) Create(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy, options ...client.CreateOption) error {
	return d.client.Create(ctx, trafficPolicy, options...)
}

func (d *trafficPolicyClient) UpdateStatus(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy, options ...client.UpdateOption) error {
	return d.client.Status().Update(ctx, trafficPolicy, options...)
}

func (d *trafficPolicyClient) List(ctx context.Context, options ...client.ListOption) (*networking_v1alpha1.TrafficPolicyList, error) {
	list := networking_v1alpha1.TrafficPolicyList{}
	err := d.client.List(ctx, &list, options...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
