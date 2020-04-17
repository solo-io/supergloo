package v1beta1

import (
	"context"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// clienset for the apiextensions.k8s.io/v1beta1 APIs
type Clientset interface {
	// clienset for the apiextensions.k8s.io/v1beta1/v1beta1 APIs
	CustomResourceDefinitions() CustomResourceDefinitionClient
}

type clientSet struct {
	client client.Client
}

func NewClientsetFromConfig(cfg *rest.Config) (*clientSet, error) {
	scheme := scheme.Scheme
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}
	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return NewClientset(client), nil
}

func NewClientset(client client.Client) *clientSet {
	return &clientSet{client: client}
}

// clienset for the apiextensions.k8s.io/v1beta1/v1beta1 APIs
func (c *clientSet) CustomResourceDefinitions() CustomResourceDefinitionClient {
	return NewCustomResourceDefinitionClient(c.client)
}

// Reader knows how to read and list CustomResourceDefinitions.
type CustomResourceDefinitionReader interface {
	// Get retrieves a CustomResourceDefinition for the given object key
	GetCustomResourceDefinition(ctx context.Context, key client.ObjectKey) (*CustomResourceDefinition, error)

	// List retrieves list of CustomResourceDefinitions for a given namespace and list options.
	ListCustomResourceDefinition(ctx context.Context, opts ...client.ListOption) (*CustomResourceDefinitionList, error)
}

// Writer knows how to create, delete, and update CustomResourceDefinitions.
type CustomResourceDefinitionWriter interface {
	// Create saves the CustomResourceDefinition object.
	CreateCustomResourceDefinition(ctx context.Context, obj *CustomResourceDefinition, opts ...client.CreateOption) error

	// Delete deletes the CustomResourceDefinition object.
	DeleteCustomResourceDefinition(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error

	// Update updates the given CustomResourceDefinition object.
	UpdateCustomResourceDefinition(ctx context.Context, obj *CustomResourceDefinition, opts ...client.UpdateOption) error

	// Patch patches the given CustomResourceDefinition object.
	PatchCustomResourceDefinition(ctx context.Context, obj *CustomResourceDefinition, patch client.Patch, opts ...client.PatchOption) error

	// DeleteAllOf deletes all CustomResourceDefinition objects matching the given options.
	DeleteAllOfCustomResourceDefinition(ctx context.Context, opts ...client.DeleteAllOfOption) error
}

// StatusWriter knows how to update status subresource of a CustomResourceDefinition object.
type CustomResourceDefinitionStatusWriter interface {
	// Update updates the fields corresponding to the status subresource for the
	// given CustomResourceDefinition object.
	UpdateCustomResourceDefinitionStatus(ctx context.Context, obj *CustomResourceDefinition, opts ...client.UpdateOption) error

	// Patch patches the given CustomResourceDefinition object's subresource.
	PatchCustomResourceDefinitionStatus(ctx context.Context, obj *CustomResourceDefinition, patch client.Patch, opts ...client.PatchOption) error
}

// Client knows how to perform CRUD operations on CustomResourceDefinitions.
type CustomResourceDefinitionClient interface {
	CustomResourceDefinitionReader
	CustomResourceDefinitionWriter
	CustomResourceDefinitionStatusWriter
}

type customResourceDefinitionClient struct {
	client client.Client
}

func NewCustomResourceDefinitionClient(client client.Client) *customResourceDefinitionClient {
	return &customResourceDefinitionClient{client: client}
}

func (c *customResourceDefinitionClient) GetCustomResourceDefinition(ctx context.Context, key client.ObjectKey) (*CustomResourceDefinition, error) {
	obj := &CustomResourceDefinition{}
	if err := c.client.Get(ctx, key, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *customResourceDefinitionClient) ListCustomResourceDefinition(ctx context.Context, opts ...client.ListOption) (*CustomResourceDefinitionList, error) {
	list := &CustomResourceDefinitionList{}
	if err := c.client.List(ctx, list, opts...); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *customResourceDefinitionClient) CreateCustomResourceDefinition(ctx context.Context, obj *CustomResourceDefinition, opts ...client.CreateOption) error {
	return c.client.Create(ctx, obj, opts...)
}

func (c *customResourceDefinitionClient) DeleteCustomResourceDefinition(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error {
	obj := &CustomResourceDefinition{}
	obj.SetName(key.Name)
	obj.SetNamespace(key.Namespace)
	return c.client.Delete(ctx, obj, opts...)
}

func (c *customResourceDefinitionClient) UpdateCustomResourceDefinition(ctx context.Context, obj *CustomResourceDefinition, opts ...client.UpdateOption) error {
	return c.client.Update(ctx, obj, opts...)
}

func (c *customResourceDefinitionClient) PatchCustomResourceDefinition(ctx context.Context, obj *CustomResourceDefinition, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *customResourceDefinitionClient) DeleteAllOfCustomResourceDefinition(ctx context.Context, opts ...client.DeleteAllOfOption) error {
	obj := &CustomResourceDefinition{}
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *customResourceDefinitionClient) UpdateCustomResourceDefinitionStatus(ctx context.Context, obj *CustomResourceDefinition, opts ...client.UpdateOption) error {
	return c.client.Status().Update(ctx, obj, opts...)
}

func (c *customResourceDefinitionClient) PatchCustomResourceDefinitionStatus(ctx context.Context, obj *CustomResourceDefinition, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Status().Patch(ctx, obj, patch, opts...)
}
