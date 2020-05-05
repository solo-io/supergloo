package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

type RestAPIDiscoveryReconciler interface {
	// Reconcile Mesh entities as indicated by the REST API with SMH's current state
	Reconcile(ctx context.Context) error
}

type RestAPIDiscoveryReconcilerFactory func(
	computeTargetName string,
	appMeshClient appmeshiface.AppMeshAPI,
	region string,
) RestAPIDiscoveryReconciler
