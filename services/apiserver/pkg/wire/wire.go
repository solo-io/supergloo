// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/server"
	multicluster_wire "github.com/solo-io/service-mesh-hub/services/common/multicluster/wire"
)

func InitializeApiServer(ctx context.Context) (ApiServerContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		managementPlaneClientsSet,
		server.NewGrpcServer,
		ApiServerContextProvider,
	)

	return ApiServerContext{}, nil
}
