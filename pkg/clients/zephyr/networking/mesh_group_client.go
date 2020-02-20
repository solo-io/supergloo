package zephyr_networking

import (
	"context"

	networkingv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshGroupClient(dynamicClient client.Client) MeshGroupClient {
	return &certificateSigningRequestClient{dynamicClient: dynamicClient}
}

type certificateSigningRequestClient struct {
	dynamicClient client.Client
}

func (c *certificateSigningRequestClient) Get(
	ctx context.Context,
	name, namespace string,
) (*networkingv1alpha1.MeshGroup, error) {

	csr := networkingv1alpha1.MeshGroup{}
	err := c.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (c *certificateSigningRequestClient) List(
	ctx context.Context,
	opts v1.ListOptions,
) (*networkingv1alpha1.MeshGroupList, error) {

	list := networkingv1alpha1.MeshGroupList{}
	err := c.dynamicClient.List(ctx, &list, &client.ListOptions{Raw: &opts})
	if err != nil {
		return nil, err
	}
	return &list, nil
}
