package helm

import (
	"context"
)

type mockHelm struct {
	create func(ctx context.Context, namespace string, manifests Manifests) error
	delete func(ctx context.Context, namespace string, manifests Manifests) error
	update func(ctx context.Context, namespace string, original, updated Manifests, recreatePods bool) error
}

func NewMockHelm(create func(ctx context.Context, namespace string, manifests Manifests) error,
	delete func(ctx context.Context, namespace string, manifests Manifests) error,
	update func(ctx context.Context, namespace string, original, updated Manifests, recreatePods bool) error) *mockHelm {
	return &mockHelm{
		create: create,
		delete: delete,
		update: update,
	}
}

func (h *mockHelm) CreateFromManifests(ctx context.Context, namespace string, manifests Manifests) error {
	return h.create(ctx, namespace, manifests)
}

func (h *mockHelm) DeleteFromManifests(ctx context.Context, namespace string, manifests Manifests) error {
	return h.delete(ctx, namespace, manifests)
}

func (h *mockHelm) UpdateFromManifests(ctx context.Context, namespace string, original, updated Manifests, recreatePods bool) error {
	return h.update(ctx, namespace, original, updated, recreatePods)
}
