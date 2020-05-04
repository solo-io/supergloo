package compute_target

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

// Callback invoked upon creation of k8s Secrets representing compute target credentials,
// which signals the initialization of a new compute target that SMH will start managing
type ComputeTargetCredentialsHandler interface {
	// Invoked when user manually registers a new compute target with SMH
	ComputeTargetAdded(ctx context.Context, secret *v1.Secret) error
	// Cleans up any state/processes associated with the specified compute target
	ComputeTargetRemoved(ctx context.Context, secret *v1.Secret) error
}
