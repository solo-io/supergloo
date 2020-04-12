package zephyr_discovery

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MeshWorkloadClientFactory func(client client.Client) MeshWorkloadClient

func MeshWorkloadClientFactoryProvider() MeshWorkloadClientFactory {
	return NewMeshWorkloadClient
}

func NewMeshWorkloadClientForConfig(cfg *rest.Config) (MeshWorkloadClient, error) {
	if err := discovery_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &meshWorkloadClient{dynamicClient: dynamicClient}, nil
}

func NewMeshWorkloadClient(client client.Client) MeshWorkloadClient {
	return &meshWorkloadClient{client}
}

type meshWorkloadClient struct {
	dynamicClient client.Client
}

func (m *meshWorkloadClient) Update(ctx context.Context, mesh *discovery_v1alpha1.MeshWorkload) error {
	return m.dynamicClient.Update(ctx, mesh)
}

func (m *meshWorkloadClient) Create(ctx context.Context, mesh *discovery_v1alpha1.MeshWorkload) error {
	return m.dynamicClient.Create(ctx, mesh)
}

func (m *meshWorkloadClient) Delete(ctx context.Context, mesh *discovery_v1alpha1.MeshWorkload) error {
	return m.dynamicClient.Delete(ctx, mesh)
}

func (m *meshWorkloadClient) Get(ctx context.Context, objKey client.ObjectKey) (*discovery_v1alpha1.MeshWorkload, error) {
	mesh := discovery_v1alpha1.MeshWorkload{}
	err := m.dynamicClient.Get(ctx, objKey, &mesh)
	if err != nil {
		return nil, err
	}
	return &mesh, nil
}

func (m *meshWorkloadClient) List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.MeshWorkloadList, error) {
	list := discovery_v1alpha1.MeshWorkloadList{}
	err := m.dynamicClient.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
