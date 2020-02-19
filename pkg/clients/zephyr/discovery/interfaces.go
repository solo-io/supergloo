package discovery_core

import (
	"context"

	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type MeshClient interface {
	Create(ctx context.Context, mesh *discoveryv1alpha1.Mesh) error
	Delete(ctx context.Context, mesh *discoveryv1alpha1.Mesh) error
	// if the mesh is not found, returns an error on which k8s.io/apimachinery/pkg/api/errors::IsNotFound should return true
	Get(ctx context.Context, objKey client.ObjectKey) (*discoveryv1alpha1.Mesh, error)
	// Will list meshes in all namespaces by default.
	// To specify a namespace call with List(ctx , client.InNamespace("namespace"))
	List(ctx context.Context, opts ...client.ListOption) (*discoveryv1alpha1.MeshList, error)
}

type MeshWorkloadClient interface {
	Create(ctx context.Context, mesh *discoveryv1alpha1.MeshWorkload) error
	Update(ctx context.Context, mesh *discoveryv1alpha1.MeshWorkload) error
	Delete(ctx context.Context, mesh *discoveryv1alpha1.MeshWorkload) error
	// if the MeshWorkload is not found, returns an error on which k8s.io/apimachinery/pkg/api/errors::IsNotFound should return true
	Get(ctx context.Context, objKey client.ObjectKey) (*discoveryv1alpha1.MeshWorkload, error)
	List(ctx context.Context, opts ...client.ListOption) (*discoveryv1alpha1.MeshWorkloadList, error)
}

// operations on the KubernetesCluster CRD
type KubernetesClusterClient interface {
	Create(ctx context.Context, cluster *discoveryv1alpha1.KubernetesCluster) error
}

type MeshServiceClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*discoveryv1alpha1.MeshService, error)
	Create(ctx context.Context, meshService *discoveryv1alpha1.MeshService, options ...client.CreateOption) error
	Update(ctx context.Context, meshService *discoveryv1alpha1.MeshService, options ...client.UpdateOption) error
}
