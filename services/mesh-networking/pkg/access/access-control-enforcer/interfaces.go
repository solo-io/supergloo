package access_policy_enforcer

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Watches events on VirtualMeshes and enforces its global access control setting
type AccessPolicyEnforcerLoop interface {
	Start(ctx context.Context) error
}

// Enforces global access control for a specific Mesh within a VirtualMesh
type AccessPolicyMeshEnforcer interface {
	// The name which will be used to identify the enforcer in the logs
	Name() string
	StartEnforcing(ctx context.Context, meshes []*discovery_v1alpha1.Mesh) error
	StopEnforcing(ctx context.Context, meshes []*discovery_v1alpha1.Mesh) error
}
