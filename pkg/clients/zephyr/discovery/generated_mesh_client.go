package zephyr_discovery

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned/typed/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type generatedMeshClient struct {
	client discoveryv1alpha1.DiscoveryV1alpha1Interface
}

func NewGeneratedMeshClient(client versioned.Interface) MeshClient {
	return &generatedMeshClient{client: client.DiscoveryV1alpha1()}
}

func (g *generatedMeshClient) Create(_ context.Context, mesh *v1alpha1.Mesh) error {
	newMesh, err := g.client.Meshes(mesh.GetNamespace()).Create(mesh)
	if err != nil {
		return err
	}
	*mesh = *newMesh
	return nil
}

func (g *generatedMeshClient) Delete(_ context.Context, mesh *v1alpha1.Mesh) error {
	return g.client.Meshes(mesh.GetNamespace()).Delete(mesh.GetName(), &v1.DeleteOptions{})
}

func (g *generatedMeshClient) Get(_ context.Context, key client.ObjectKey) (*v1alpha1.Mesh, error) {
	return g.client.Meshes(key.Namespace).Get(key.Name, v1.GetOptions{})
}

func (g *generatedMeshClient) List(_ context.Context, opts ...client.ListOption) (*v1alpha1.MeshList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range opts {
		v.ApplyToList(listOptions)
	}
	raw := v1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return g.client.Meshes("").List(raw)
}

func (g *generatedMeshClient) UpsertSpec(ctx context.Context, mesh *v1alpha1.Mesh) error {
	key := client.ObjectKey{Name: mesh.GetName(), Namespace: mesh.GetNamespace()}
	existingMesh, err := g.Get(ctx, key)
	if err != nil {
		if errors.IsNotFound(err) {
			return g.Create(ctx, mesh)
		}
		return err
	}
	existingMesh.Spec = mesh.Spec
	return g.Update(ctx, existingMesh)
}

func (g *generatedMeshClient) Update(_ context.Context, mesh *v1alpha1.Mesh) error {
	newMesh, err := g.client.Meshes(mesh.GetNamespace()).Update(mesh)
	if err != nil {
		return err
	}
	*mesh = *newMesh
	return nil
}
