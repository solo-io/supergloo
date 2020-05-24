package crd_uninstall

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type CrdRemover interface {
	// Remove all registered zephyr.solo.io CRDs. The cluster name is just used for error logging purposes
	// returns:
	//  - (true, nil) if CRDs were found and deleted
	//  - (false, nil) if no CRDs were found
	//  - (_, err) if the operation failed
	RemoveZephyrCrds(ctx context.Context, clusterName string, remoteKubeConfig *rest.Config) (crdsDeleted bool, err error)

	// Remove just the CRDs belonging to the group in the given `groupVersion`
	// Similar return semantics as the above method
	RemoveCrdGroup(ctx context.Context, clusterName string, remoteKubeConfig *rest.Config, groupVersion schema.GroupVersion) (crdsDeleted bool, err error)
}
