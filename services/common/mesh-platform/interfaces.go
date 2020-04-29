package mesh_platform

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

// Callback invoked upon creation of k8s Secrets representing MeshPlatform credentials,
// which signals the initialization of a new MeshPlatform that SMH will start managing
type MeshPlatformCredentialsHandler interface {
	// Invoked when user manually registers a new service mesh platform with SMH
	MeshPlatformAdded(ctx context.Context, secret *v1.Secret) error
	// Cleans up any state/processes associated with the specified MeshPlatform
	MeshPlatformRemoved(ctx context.Context, secret *v1.Secret) error
}
