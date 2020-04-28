package rest

import (
	"context"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

type RestAPIDiscoveryReconciler interface {
	// Reconcile Mesh entities as indicated by the REST API with SMH's current state
	Reconcile(ctx context.Context) error
}
