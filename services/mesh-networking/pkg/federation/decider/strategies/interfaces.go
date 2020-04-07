package strategies

import (
	"context"

	"github.com/rotisserie/eris"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
)

var (
	UnsupportedMode = func(mode networking_types.VirtualMeshSpec_Federation_Mode) error {
		return eris.Errorf("Mode %+v is not supported", mode)
	}
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type FederationStrategy interface {
	WriteFederationToServices(
		ctx context.Context,
		vm *networking_v1alpha1.VirtualMesh,
		meshNameToMetadata MeshNameToMetadata,
	) error
}

type FederationStrategyChooser func(
	mode networking_types.VirtualMeshSpec_Federation_Mode,
	meshServiceClient discovery_core.MeshServiceClient,
) (FederationStrategy, error)

func NewFederationStrategyChooser() FederationStrategyChooser {
	return GetFederationStrategyFromMode
}

var GetFederationStrategyFromMode FederationStrategyChooser = func(
	mode networking_types.VirtualMeshSpec_Federation_Mode,
	meshServiceClient discovery_core.MeshServiceClient,
) (FederationStrategy, error) {

	switch mode {
	case networking_types.VirtualMeshSpec_Federation_PERMISSIVE:
		return NewPermissiveFederation(meshServiceClient), nil
	default:
		return nil, UnsupportedMode(mode)
	}
}
