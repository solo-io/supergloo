package zephyr_networking

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/clientset/versioned"
	networkingv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/clientset/versioned/typed/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type generatedVirtualMeshClient struct {
	client networkingv1alpha1.NetworkingV1alpha1Interface
}

func NewGeneratedVirtualMeshClient(cfg *rest.Config) VirtualMeshClient {
	clientSet, _ := versioned.NewForConfig(cfg)
	return &generatedVirtualMeshClient{client: clientSet.NetworkingV1alpha1()}
}

func (g *generatedVirtualMeshClient) Get(_ context.Context, name, namespace string) (*v1alpha1.VirtualMesh, error) {
	return g.client.VirtualMeshes(namespace).Get(name, v1.GetOptions{})
}

func (g *generatedVirtualMeshClient) List(_ context.Context, opts ...client.ListOption) (*v1alpha1.VirtualMeshList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range opts {
		v.ApplyToList(listOptions)
	}
	raw := v1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return g.client.VirtualMeshes(listOptions.Namespace).List(raw)
}

func (g *generatedVirtualMeshClient) UpdateStatus(_ context.Context, virtualMesh *v1alpha1.VirtualMesh, _ ...client.UpdateOption) error {
	_, err := g.client.VirtualMeshes(virtualMesh.GetNamespace()).UpdateStatus(virtualMesh)
	return err
}

func (g *generatedVirtualMeshClient) Create(ctx context.Context, virtualMesh *v1alpha1.VirtualMesh) error {
	newVirtualMesh, err := g.client.VirtualMeshes(virtualMesh.GetNamespace()).Create(virtualMesh)
	if err != nil {
		return err
	}
	*virtualMesh = *newVirtualMesh
	return nil
}
