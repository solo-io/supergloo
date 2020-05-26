package appmesh

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshAccessControlDao interface {
	ListMeshServicesForMesh(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) ([]*zephyr_discovery.MeshService, error)

	ListMeshWorkloadsForMesh(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) ([]*zephyr_discovery.MeshWorkload, error)

	// For a given MeshService, creates an Appmesh VirtualService if one does not already exist. Return the VirtualServiceRef.
	EnsureAppmeshVirtualService(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		meshService *zephyr_discovery.MeshService,
	) (appmesh.VirtualServiceRef, error)

	// For a given MeshWorkload, creates an Appmesh VirtualNode if one does not already exist, and that
	// all VirtualServices are declared as backends on the VirtualNode.
	EnsureVirtualNodeBackends(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		meshWorkload *zephyr_discovery.MeshWorkload,
		virtualServiceNames []string,
	) error
}
