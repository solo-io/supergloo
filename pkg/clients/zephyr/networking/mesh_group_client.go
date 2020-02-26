package zephyr_networking

import (
	"context"

	networkingv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshGroupClient(dynamicClient client.Client) MeshGroupClient {
	return &meshGroupClient{dynamicClient: dynamicClient}
}

type meshGroupClient struct {
	dynamicClient client.Client
}

func (m *meshGroupClient) Get(
	ctx context.Context,
	name, namespace string,
) (*networkingv1alpha1.MeshGroup, error) {

	csr := networkingv1alpha1.MeshGroup{}
	err := m.dynamicClient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, &csr)
	if err != nil {
		return nil, err
	}
	return &csr, nil
}

func (m *meshGroupClient) List(
	ctx context.Context,
	opts v1.ListOptions,
) (*networkingv1alpha1.MeshGroupList, error) {

	list := networkingv1alpha1.MeshGroupList{}
	err := m.dynamicClient.List(ctx, &list, &client.ListOptions{Raw: &opts})
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (m *meshGroupClient) UpdateStatus(ctx context.Context, meshGroup *networkingv1alpha1.MeshGroup, opts ...client.UpdateOption) error {
	return m.dynamicClient.Status().Update(ctx, meshGroup, opts...)
}
