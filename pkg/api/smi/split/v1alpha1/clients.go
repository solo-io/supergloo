package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
)

// clienset for the split/v1alpha1 APIs
type Clientset interface {
	// clienset for the split/v1alpha1/v1alpha1 APIs
	TrafficSplits() TrafficSplitClient
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

// clienset for the split/v1alpha1/v1alpha1 APIs
func (c *clientSet) TrafficSplits() TrafficSplitClient {
	return NewTrafficSplitClient(c.client)
}

// Reader knows how to read and list TrafficSplits.
type TrafficSplitReader interface {
	// Get retrieves a TrafficSplit for the given object key
	GetTrafficSplit(ctx context.Context, key client.ObjectKey) (*TrafficSplit, error)

	// List retrieves list of TrafficSplits for a given namespace and list options.
	ListTrafficSplit(ctx context.Context, opts ...client.ListOption) (*TrafficSplitList, error)
}

// Writer knows how to create, delete, and update TrafficSplits.
type TrafficSplitWriter interface {
	// Create saves the TrafficSplit object.
	CreateTrafficSplit(ctx context.Context, obj *TrafficSplit, opts ...client.CreateOption) error

	// Delete deletes the TrafficSplit object.
	DeleteTrafficSplit(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error

	// Update updates the given TrafficSplit object.
	UpdateTrafficSplit(ctx context.Context, obj *TrafficSplit, opts ...client.UpdateOption) error

	// If the TrafficSplit object exists, update its spec. Otherwise, create the TrafficSplit object.
	UpsertTrafficSplitSpec(ctx context.Context, obj *TrafficSplit, opts ...client.UpdateOption) error

	// Patch patches the given TrafficSplit object.
	PatchTrafficSplit(ctx context.Context, obj *TrafficSplit, patch client.Patch, opts ...client.PatchOption) error

	// DeleteAllOf deletes all TrafficSplit objects matching the given options.
	DeleteAllOfTrafficSplit(ctx context.Context, opts ...client.DeleteAllOfOption) error
}

// StatusWriter knows how to update status subresource of a TrafficSplit object.
type TrafficSplitStatusWriter interface {
	// Update updates the fields corresponding to the status subresource for the
	// given TrafficSplit object.
	UpdateTrafficSplitStatus(ctx context.Context, obj *TrafficSplit, opts ...client.UpdateOption) error

	// Patch patches the given TrafficSplit object's subresource.
	PatchTrafficSplitStatus(ctx context.Context, obj *TrafficSplit, patch client.Patch, opts ...client.PatchOption) error
}

// Client knows how to perform CRUD operations on TrafficSplits.
type TrafficSplitClient interface {
	TrafficSplitReader
	TrafficSplitWriter
	TrafficSplitStatusWriter
}

type trafficSplitClient struct {
	client client.Client
}

func NewTrafficSplitClient(client client.Client) *trafficSplitClient {
	return &trafficSplitClient{client: client}
}

func (c *trafficSplitClient) GetTrafficSplit(ctx context.Context, key client.ObjectKey) (*TrafficSplit, error) {
	obj := &TrafficSplit{}
	if err := c.client.Get(ctx, key, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *trafficSplitClient) ListTrafficSplit(ctx context.Context, opts ...client.ListOption) (*TrafficSplitList, error) {
	list := &TrafficSplitList{}
	if err := c.client.List(ctx, list, opts...); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *trafficSplitClient) CreateTrafficSplit(ctx context.Context, obj *TrafficSplit, opts ...client.CreateOption) error {
	return c.client.Create(ctx, obj, opts...)
}

func (c *trafficSplitClient) DeleteTrafficSplit(ctx context.Context, key client.ObjectKey, opts ...client.DeleteOption) error {
	obj := &TrafficSplit{}
	obj.SetName(key.Name)
	obj.SetNamespace(key.Namespace)
	return c.client.Delete(ctx, obj, opts...)
}

func (c *trafficSplitClient) UpdateTrafficSplit(ctx context.Context, obj *TrafficSplit, opts ...client.UpdateOption) error {
	return c.client.Update(ctx, obj, opts...)
}

func (c *trafficSplitClient) UpsertTrafficSplitSpec(ctx context.Context, obj *TrafficSplit, opts ...client.UpdateOption) error {
	existing, err := c.GetTrafficSplit(ctx, client.ObjectKey{Name: obj.GetName(), Namespace: obj.GetNamespace()})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.CreateTrafficSplit(ctx, obj)
		}
		return err
	}
	existing.Spec = obj.Spec
	return c.client.Update(ctx, existing, opts...)
}

func (c *trafficSplitClient) PatchTrafficSplit(ctx context.Context, obj *TrafficSplit, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *trafficSplitClient) DeleteAllOfTrafficSplit(ctx context.Context, opts ...client.DeleteAllOfOption) error {
	obj := &TrafficSplit{}
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *trafficSplitClient) UpdateTrafficSplitStatus(ctx context.Context, obj *TrafficSplit, opts ...client.UpdateOption) error {
	return c.client.Status().Update(ctx, obj, opts...)
}

func (c *trafficSplitClient) PatchTrafficSplitStatus(ctx context.Context, obj *TrafficSplit, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Status().Patch(ctx, obj, patch, opts...)
}
