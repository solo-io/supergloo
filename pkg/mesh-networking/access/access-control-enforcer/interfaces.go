package access_policy_enforcer

import (
	"context"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Watches events on VirtualMeshes and enforces its global access control setting
type AccessPolicyEnforcerLoop interface {
	Start(ctx context.Context) error
}
