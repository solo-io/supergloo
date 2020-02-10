//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	multicluster_wire "github.com/solo-io/mesh-projects/services/common/multicluster/wire"
	"github.com/solo-io/mesh-projects/services/mesh-group/pkg/controller"
)

func InitializeMeshGroup(ctx context.Context) (MeshGroupContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		controller.MeshGroupProviderSet,
		MeshGroupContextProvider,
	)

	return MeshGroupContext{}, nil
}
