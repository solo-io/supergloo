package zephyr_discovery

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshClientFactoryProvider() MeshClientFactory {
	return NewMeshClient
}

func NewMeshClientForConfig(cfg *rest.Config) (MeshClient, error) {
	if err := discovery_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &meshClient{client: dynamicClient}, nil
}

func NewMeshClient(client client.Client) MeshClient {
	return &meshClient{client: client}
}

type meshClient struct {
	client client.Client
}

func (m *meshClient) Create(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error {
	return m.client.Create(ctx, mesh)
}

func (m *meshClient) Delete(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error {
	return m.client.Delete(ctx, mesh)
}

func (m *meshClient) Get(ctx context.Context, objKey client.ObjectKey) (*discovery_v1alpha1.Mesh, error) {
	mesh := discovery_v1alpha1.Mesh{}
	err := m.client.Get(ctx, objKey, &mesh)

	if err != nil {
		return nil, err
	}
	return &mesh, nil
}

func (m *meshClient) List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.MeshList, error) {
	result := discovery_v1alpha1.MeshList{}
	err := m.client.List(ctx, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *meshClient) Update(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error {
	return m.client.Update(ctx, mesh)
}

func (m *meshClient) UpsertSpec(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error {
	key := client.ObjectKey{Name: mesh.GetName(), Namespace: mesh.GetNamespace()}
	existingMesh, err := m.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return m.Create(ctx, mesh)
		}
		return err
	}
	existingMesh.Spec = mesh.Spec
	return m.Update(ctx, existingMesh)
}
