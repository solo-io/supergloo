package access_control_enforcer

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

/*
	Enforces global access control for a specific Mesh within a VirtualMesh depending on
	the value of "enforce_access_control" on the VirtualMesh.
	If the VirtualMesh is nil, enforce the mesh specific default.
*/
type AccessPolicyMeshEnforcer interface {
	// The name which will be used to identify the enforcer in the logs
	Name() string

	// Reconcile mesh specific resources to reflect the intended access control state as declared in the VirtualMesh.
	ReconcileAccessControl(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		virtualMesh *zephyr_networking.VirtualMesh,
	) error
}
