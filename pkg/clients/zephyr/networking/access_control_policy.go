package zephyr_networking

import (
	"context"

	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type accessControlPolicyClient struct {
	client client.Client
}

func NewAccessControlPolicyClientForConfig(cfg *rest.Config) (AccessControlPolicyClient, error) {
	if err := networking_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &accessControlPolicyClient{client: dynamicClient}, nil
}

func NewAccessControlPolicyClient(client client.Client) AccessControlPolicyClient {
	return &accessControlPolicyClient{client: client}
}

func (a *accessControlPolicyClient) List(ctx context.Context, opts ...client.ListOption) (*networking_v1alpha1.AccessControlPolicyList, error) {
	list := networking_v1alpha1.AccessControlPolicyList{}
	err := a.client.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (a *accessControlPolicyClient) UpdateStatus(ctx context.Context, acp *networking_v1alpha1.AccessControlPolicy, options ...client.UpdateOption) error {
	return a.client.Status().Update(ctx, acp, options...)
}

func (a *accessControlPolicyClient) Create(ctx context.Context, acp *networking_v1alpha1.AccessControlPolicy, options ...client.CreateOption) error {
	return a.client.Create(ctx, acp, options...)
}
