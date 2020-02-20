//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	multicluster_wire "github.com/solo-io/mesh-projects/services/common/multicluster/wire"
	group_controller "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/groups/controller"
	networking_multicluster "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster"
)

func InitializeMeshNetworking(ctx context.Context) (MeshNetworkingContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		group_controller.MeshGroupProviderSet,
		networking_multicluster.NewCSRControllerFactory,
		networking_multicluster.NewMeshNetworkingClusterHandler,
		MeshNetworkingContextProvider,
	)

	return MeshNetworkingContext{}, nil
}
