package discovery_core

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshWorkloadClient(client client.Client) MeshWorkloadClient {
	return &meshWorkloadClient{client}
}

type meshWorkloadClient struct {
	kubeClient client.Client
}

func (m *meshWorkloadClient) Create(ctx context.Context, mesh *v1alpha1.MeshWorkload) error {
	return m.kubeClient.Create(ctx, mesh)
}

func (m *meshWorkloadClient) Delete(ctx context.Context, mesh *v1alpha1.MeshWorkload) error {
	return m.kubeClient.Delete(ctx, mesh)
}

func (m *meshWorkloadClient) Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha1.MeshWorkload, error) {
	mesh := v1alpha1.MeshWorkload{}
	err := m.kubeClient.Get(ctx, objKey, &mesh)
	if err != nil {
		return nil, err
	}
	return &mesh, nil
}
