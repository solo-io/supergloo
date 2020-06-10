package internal_watcher

import (
	"context"

	k8s_core_types "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/membership.go

// Handles creation/removal of k8s Secrets representing credentials for compute targets to be managed by SMH
type ComputeTargetSecretHandler interface {
	ComputeTargetSecretAdded(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error)
	ComputeTargetSecretRemoved(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error)
}
