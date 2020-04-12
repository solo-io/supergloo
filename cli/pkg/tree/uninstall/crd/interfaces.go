package crd_uninstall

import (
	"context"

	"k8s.io/client-go/rest"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type CrdRemover interface {
	// the cluster name is just used for error logging purposes
	// returns:
	//  - (true, nil) if CRDs were found and deleted
	//  - (false, nil) if no CRDs were found
	//  - (_, err) if the operation failed
	RemoveZephyrCrds(ctx context.Context, clusterName string, remoteKubeConfig *rest.Config) (crdsDeleted bool, err error)
}
