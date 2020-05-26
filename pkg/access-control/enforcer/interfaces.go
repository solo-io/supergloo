package access_control_enforcer

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Enforces global access control for a specific Mesh within a VirtualMesh
type AccessPolicyMeshEnforcer interface {
	// The name which will be used to identify the enforcer in the logs
	Name() string
	StartEnforcing(ctx context.Context, mesh *zephyr_discovery.Mesh) error
	StopEnforcing(ctx context.Context, mesh *zephyr_discovery.Mesh) error
}
