package zephyr_discovery

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type MeshClientFactory func(client client.Client) MeshClient

type MeshClient interface {
	Create(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error
	Delete(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error
	// if the mesh is not found, returns an error on which k8s.io/apimachinery/pkg/api/errors::IsNotFound should return true
	Get(ctx context.Context, objKey client.ObjectKey) (*discovery_v1alpha1.Mesh, error)
	// Will list meshes in all namespaces by default.
	// To specify a namespace call with List(ctx , client.InNamespace("namespace"))
	List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.MeshList, error)
	// Create if Mesh doesn't exist, otherwise update.
	UpsertSpec(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error
	Update(ctx context.Context, mesh *discovery_v1alpha1.Mesh) error
}

type MeshWorkloadClient interface {
	Create(ctx context.Context, meshWorkload *discovery_v1alpha1.MeshWorkload) error
	Update(ctx context.Context, meshWorkload *discovery_v1alpha1.MeshWorkload) error
	Delete(ctx context.Context, meshWorkload *discovery_v1alpha1.MeshWorkload) error
	// if the MeshWorkload is not found, returns an error on which k8s.io/apimachinery/pkg/api/errors::IsNotFound should return true
	Get(ctx context.Context, objKey client.ObjectKey) (*discovery_v1alpha1.MeshWorkload, error)
	List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.MeshWorkloadList, error)
}

// operations on the KubernetesCluster CRD
type KubernetesClusterClient interface {
	Create(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error
	Update(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error
	Upsert(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error
	Get(ctx context.Context, key client.ObjectKey) (*discovery_v1alpha1.KubernetesCluster, error)
	List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.KubernetesClusterList, error)
}

type MeshServiceClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*discovery_v1alpha1.MeshService, error)
	Create(ctx context.Context, meshService *discovery_v1alpha1.MeshService, options ...client.CreateOption) error
	Update(ctx context.Context, meshService *discovery_v1alpha1.MeshService, options ...client.UpdateOption) error
	UpdateStatus(ctx context.Context, meshService *discovery_v1alpha1.MeshService, options ...client.UpdateOption) error
	List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.MeshServiceList, error)
}
