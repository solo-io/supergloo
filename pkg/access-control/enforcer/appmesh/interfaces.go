package appmesh

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshAccessControlDao interface {
	// Return two maps which associate workloads to services and vice versa.
	GetServicesAndWorkloadsForMesh(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) (map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload,
		map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService,
		error)

	// For a given MeshService, creates an Appmesh VirtualService if one does not already exist,
	// backed by a VirtualRouter provider with a set of default routes to all workloads backing the service.
	// Return the VirtualService name.
	EnsureVirtualServicesWithDefaultRoutes(
		mesh *zephyr_discovery.Mesh,
		serviceToWorkloads map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload,
	) error

	// For a given MeshWorkload, creates an Appmesh VirtualNode if one does not already exist, and that
	// all VirtualServices are declared as backends on the VirtualNode.
	EnsureVirtualNodesWithDefaultBackends(
		mesh *zephyr_discovery.Mesh,
		workloadToServices map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService,
	) error
}
