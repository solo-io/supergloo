package strategies

import (
	"context"

	"github.com/rotisserie/eris"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
)

var (
	UnsupportedMode = func(mode networking_types.Federation_Mode) error {
		return eris.Errorf("Mode %+v is not supported", mode)
	}
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type FederationStrategy interface {
	WriteFederationToServices(
		ctx context.Context,
		group *networking_v1alpha1.MeshGroup,
		meshNameToMetadata MeshNameToMetadata,
	) error
}

type FederationStrategyChooser func(mode networking_types.Federation_Mode, meshServiceClient discovery_core.MeshServiceClient) (FederationStrategy, error)

var GetFederationStrategyFromMode FederationStrategyChooser = func(
	mode networking_types.Federation_Mode,
	meshServiceClient discovery_core.MeshServiceClient,
) (FederationStrategy, error) {

	switch mode {
	case networking_types.Federation_PERMISSIVE:
		return NewPermissiveFederation(meshServiceClient), nil
	default:
		return nil, UnsupportedMode(mode)
	}
}
