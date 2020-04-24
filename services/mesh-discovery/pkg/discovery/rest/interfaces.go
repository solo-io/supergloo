package rest

import (
	"context"

	k8s_core_types "k8s.io/api/core/v1"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

// Callback invoked upon creation of k8s Secrets representing REST API credentials, which signals the initialization of a new REST API
// that SMH should start managing
type RestAPICredsHandler interface {
	// Invoked when user manually registers a new service-mesh REST API with SMH, spawns a RestAPIDiscoveryReconciler and periodically reconcile
	RestAPIAdded(parentCtx context.Context, secret *k8s_core_types.Secret) error
	// Cleans up any state/processes associated with the given REST API
	RestAPIRemoved(ctx context.Context, secret *k8s_core_types.Secret) error
}

type RestAPIDiscoveryReconciler interface {
	// Reconcile Mesh entities as indicated by the REST API with SMH's current state
	Reconcile(ctx context.Context) error
}
