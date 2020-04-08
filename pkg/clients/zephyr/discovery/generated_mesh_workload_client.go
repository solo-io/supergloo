package zephyr_discovery

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned/typed/discovery.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type generatedMeshWorkloadClient struct {
	client discoveryv1alpha1.DiscoveryV1alpha1Interface
}

func NewGeneratedMeshWorkloadClient(client versioned.Interface) MeshWorkloadClient {
	return &generatedMeshWorkloadClient{client: client.DiscoveryV1alpha1()}
}

func (g *generatedMeshWorkloadClient) Create(_ context.Context, meshWorkload *v1alpha1.MeshWorkload) error {
	newMeshWorkload, err := g.client.MeshWorkloads(meshWorkload.GetNamespace()).Create(meshWorkload)
	if err != nil {
		return err
	}
	*meshWorkload = *newMeshWorkload
	return nil
}

func (g *generatedMeshWorkloadClient) Update(_ context.Context, meshWorkload *v1alpha1.MeshWorkload) error {
	newMesh, err := g.client.MeshWorkloads(meshWorkload.GetNamespace()).Update(meshWorkload)
	if err != nil {
		return err
	}
	*meshWorkload = *newMesh
	return nil
}

func (g *generatedMeshWorkloadClient) Delete(_ context.Context, meshWorkload *v1alpha1.MeshWorkload) error {
	return g.client.MeshWorkloads(meshWorkload.GetNamespace()).Delete(meshWorkload.GetName(), &v1.DeleteOptions{})
}

func (g *generatedMeshWorkloadClient) Get(_ context.Context, objKey client.ObjectKey) (*v1alpha1.MeshWorkload, error) {
	return g.client.MeshWorkloads(objKey.Namespace).Get(objKey.Name, v1.GetOptions{})
}

func (g *generatedMeshWorkloadClient) List(_ context.Context, opts ...client.ListOption) (*v1alpha1.MeshWorkloadList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range opts {
		v.ApplyToList(listOptions)
	}
	raw := v1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return g.client.MeshWorkloads("").List(raw)
}
