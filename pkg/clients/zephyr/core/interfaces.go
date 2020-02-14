package zephyr_core

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type MeshClient interface {
	Create(ctx context.Context, mesh *v1alpha1.Mesh) error
	Delete(ctx context.Context, mesh *v1alpha1.Mesh) error
	// if the mesh is not found, returns an error on which k8s.io/apimachinery/pkg/api/errors::IsNotFound should return true
	Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha1.Mesh, error)
	// Will list meshes in all namespaces by default.
	// To specify a namespace call with List(ctx , client.InNamespace("namespace"))
	List(ctx context.Context, opts ...client.ListOption) (*v1alpha1.MeshList, error)
}

type MeshWorkloadClient interface {
	Create(ctx context.Context, mesh *v1alpha1.MeshWorkload) error
	Update(ctx context.Context, mesh *v1alpha1.MeshWorkload) error
	Delete(ctx context.Context, mesh *v1alpha1.MeshWorkload) error
	// if the MeshWorkload is not found, returns an error on which k8s.io/apimachinery/pkg/api/errors::IsNotFound should return true
	Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha1.MeshWorkload, error)
}
