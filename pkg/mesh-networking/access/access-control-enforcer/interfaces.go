package access_policy_enforcer

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Watches events on VirtualMeshes and enforces its global access control setting
type AccessPolicyEnforcerLoop interface {
	Start(ctx context.Context) error
}

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
		mesh *v1alpha1.Mesh,
		virtualMesh *v1alpha12.VirtualMesh,
	) error
}
