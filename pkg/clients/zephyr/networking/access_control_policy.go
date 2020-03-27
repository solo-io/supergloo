package zephyr_networking

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type accessControlPolicyClient struct {
	client client.Client
}

func NewAccessControlPolicyClient(client client.Client) AccessControlPolicyClient {
	return &accessControlPolicyClient{client: client}
}

func (a *accessControlPolicyClient) List(ctx context.Context, opts ...client.ListOption) (*v1alpha1.AccessControlPolicyList, error) {
	list := v1alpha1.AccessControlPolicyList{}
	err := a.client.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (a *accessControlPolicyClient) UpdateStatus(ctx context.Context, acp *v1alpha1.AccessControlPolicy, options ...client.UpdateOption) error {
	return a.client.Status().Update(ctx, acp, options...)
}
