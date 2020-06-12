package aws

import (
	"context"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

type RestAPIDiscoveryReconciler interface {
	// Reconcile Mesh entities as indicated by the REST API with SMH's current state
	Reconcile(ctx context.Context, accountID string) error
	GetName() string
}

type EksDiscoveryReconciler RestAPIDiscoveryReconciler

type AppMeshDiscoveryReconciler RestAPIDiscoveryReconciler
