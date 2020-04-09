// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	multicluster_wire "github.com/solo-io/service-mesh-hub/services/common/multicluster/wire"
)

func InitializeApiServer(ctx context.Context) (ApiServerContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		ApiServerContextProvider,
	)

	return ApiServerContext{}, nil
}
