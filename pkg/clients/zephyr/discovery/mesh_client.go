package discovery_core

import (
	"context"

	mp_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshClient(mgr mc_manager.AsyncManager) MeshClient {
	return &meshClient{mgr.Manager().GetClient()}
}

type meshClient struct {
	client client.Client
}

func (m *meshClient) Create(ctx context.Context, mesh *mp_v1alpha1.Mesh) error {
	return m.client.Create(ctx, mesh)
}

func (m *meshClient) Delete(ctx context.Context, mesh *mp_v1alpha1.Mesh) error {
	return m.client.Delete(ctx, mesh)
}

func (m *meshClient) Get(ctx context.Context, objKey client.ObjectKey) (*mp_v1alpha1.Mesh, error) {
	mesh := mp_v1alpha1.Mesh{}
	err := m.client.Get(ctx, objKey, &mesh)

	if err != nil {
		return nil, err
	}
	return &mesh, nil
}

func (m *meshClient) List(ctx context.Context, opts ...client.ListOption) (*mp_v1alpha1.MeshList, error) {
	result := mp_v1alpha1.MeshList{}
	err := m.client.List(ctx, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
