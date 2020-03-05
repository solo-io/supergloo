package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	group_controller "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/groups/controller"
)

var (
	MeshGroupProviderSet = wire.NewSet(
		zephyr_core.NewMeshClient,
		MeshGroupValidatorProvider,
		MeshGroupEventHandlerProvider,
	)
)

func MeshGroupEventHandlerProvider(ctx context.Context, validator group_controller.MeshGroupValidator) controller.MeshGroupEventHandler {
	return group_controller.NewMeshGroupEventHandler(ctx, validator)
}

func MeshGroupValidatorProvider(meshClient zephyr_core.MeshClient) group_controller.MeshGroupValidator {
	return group_controller.NewMeshGroupValidator(meshClient)
}
