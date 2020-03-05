package zephyr_discovery

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MeshServiceClientFactory func(client client.Client) MeshServiceClient

func MeshServiceClientFactoryProvider() MeshServiceClientFactory {
	return NewMeshServiceClient
}

func NewMeshServiceClient(client client.Client) MeshServiceClient {
	return &meshServiceClient{
		client: client,
	}
}

type meshServiceClient struct {
	client client.Client
}

func (m *meshServiceClient) Get(ctx context.Context, key client.ObjectKey) (*v1alpha1.MeshService, error) {
	meshService := v1alpha1.MeshService{}
	err := m.client.Get(ctx, key, &meshService)
	if err != nil {
		return nil, err
	}

	return &meshService, nil
}

func (m *meshServiceClient) Create(ctx context.Context, meshService *v1alpha1.MeshService, options ...client.CreateOption) error {
	return m.client.Create(ctx, meshService, options...)
}

func (m *meshServiceClient) Update(ctx context.Context, meshService *v1alpha1.MeshService, options ...client.UpdateOption) error {
	return m.client.Update(ctx, meshService, options...)
}

func (m *meshServiceClient) List(ctx context.Context, opts ...client.ListOption) (*v1alpha1.MeshServiceList, error) {
	list := v1alpha1.MeshServiceList{}
	err := m.client.List(ctx, &list, opts...)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (m *meshServiceClient) UpdateStatus(ctx context.Context, meshService *v1alpha1.MeshService, options ...client.UpdateOption) error {
	return m.client.Status().Update(ctx, meshService, options...)
}
