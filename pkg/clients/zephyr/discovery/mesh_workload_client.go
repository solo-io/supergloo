package zephyr_discovery

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MeshWorkloadClientFactory func(client client.Client) MeshWorkloadClient

func MeshWorkloadClientFactoryProvider() MeshWorkloadClientFactory {
	return NewMeshWorkloadClient
}

func NewMeshWorkloadClient(client client.Client) MeshWorkloadClient {
	return &meshWorkloadClient{client}
}

type meshWorkloadClient struct {
	dynamicClient client.Client
}

func (m *meshWorkloadClient) Update(ctx context.Context, mesh *v1alpha1.MeshWorkload) error {
	return m.dynamicClient.Update(ctx, mesh)
}

func (m *meshWorkloadClient) Create(ctx context.Context, mesh *v1alpha1.MeshWorkload) error {
	return m.dynamicClient.Create(ctx, mesh)
}

func (m *meshWorkloadClient) Delete(ctx context.Context, mesh *v1alpha1.MeshWorkload) error {
	return m.dynamicClient.Delete(ctx, mesh)
}

func (m *meshWorkloadClient) Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha1.MeshWorkload, error) {
	mesh := v1alpha1.MeshWorkload{}
	err := m.dynamicClient.Get(ctx, objKey, &mesh)
	if err != nil {
		return nil, err
	}
	return &mesh, nil
}

func (m *meshWorkloadClient) List(ctx context.Context, opts ...client.ListOption) (*v1alpha1.MeshWorkloadList, error) {
	list := v1alpha1.MeshWorkloadList{}
	err := m.dynamicClient.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
