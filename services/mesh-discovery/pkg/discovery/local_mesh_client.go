package discovery

import (
	"context"

	mp_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./local_mesh_client.go -destination mocks/mock_local_mesh_client.go

type LocalMeshClient interface {
	Create(ctx context.Context, mesh *mp_v1alpha1.Mesh) error
	Delete(ctx context.Context, mesh *mp_v1alpha1.Mesh) error

	// if the mesh is not found, returns an error on which k8s.io/apimachinery/pkg/api/errors::IsNotFound should return true
	Get(ctx context.Context, objKey client.ObjectKey) (*mp_v1alpha1.Mesh, error)
}

func NewLocalMeshClient(localManager mc_manager.AsyncManager) LocalMeshClient {
	return &localMeshClient{localManager.Manager().GetClient()}
}

type localMeshClient struct {
	localClient client.Client
}

func (m *localMeshClient) Create(ctx context.Context, mesh *mp_v1alpha1.Mesh) error {
	return m.localClient.Create(ctx, mesh)
}

func (m *localMeshClient) Delete(ctx context.Context, mesh *mp_v1alpha1.Mesh) error {
	return m.localClient.Delete(ctx, mesh)
}

func (m *localMeshClient) Get(ctx context.Context, objKey client.ObjectKey) (*mp_v1alpha1.Mesh, error) {
	mesh := mp_v1alpha1.Mesh{}
	err := m.localClient.Get(ctx, objKey, &mesh)

	if err != nil {
		return nil, err
	}
	return &mesh, nil
}
