package zephyr_discovery

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned/typed/discovery.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func NewGeneratedMeshServiceClient(disc versioned.Interface) MeshServiceClient {
	return &generatedMeshServiceClient{
		client: disc.DiscoveryV1alpha1(),
	}
}

type generatedMeshServiceClient struct {
	client discoveryv1alpha1.MeshServicesGetter
}

func (m *generatedMeshServiceClient) Get(ctx context.Context, key client.ObjectKey) (*v1alpha1.MeshService, error) {
	return m.client.MeshServices(key.Namespace).Get(key.Name, v1.GetOptions{})
}

func (m *generatedMeshServiceClient) Create(ctx context.Context, meshService *v1alpha1.MeshService, options ...client.CreateOption) error {
	newMeshService, err := m.client.MeshServices(meshService.GetNamespace()).Create(meshService)
	if err != nil {
		return err
	}
	*meshService = *newMeshService
	return nil
}

func (m *generatedMeshServiceClient) Update(ctx context.Context, meshService *v1alpha1.MeshService, options ...client.UpdateOption) error {
	newMeshService, err := m.client.MeshServices(meshService.GetNamespace()).Update(meshService)
	if err != nil {
		return err
	}
	*meshService = *newMeshService
	return nil
}

func (m *generatedMeshServiceClient) List(ctx context.Context, opts ...client.ListOption) (*v1alpha1.MeshServiceList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range opts {
		v.ApplyToList(listOptions)
	}
	raw := v1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return m.client.MeshServices(listOptions.Namespace).List(raw)
}

func (m *generatedMeshServiceClient) UpdateStatus(ctx context.Context, meshService *v1alpha1.MeshService, options ...client.UpdateOption) error {
	_, err := m.client.MeshServices(meshService.GetNamespace()).UpdateStatus(meshService)
	return err
}
