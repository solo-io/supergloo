// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/handlers"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/server"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/server/health_check"
	multicluster_wire "github.com/solo-io/service-mesh-hub/services/common/multicluster/wire"
)

func InitializeApiServer(ctx context.Context) (ApiServerContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		managementPlaneClientsSet,
		handlers.HandlerSet,
		health_check.NewHealthChecker,
		server.NewGrpcServer,
		ApiServerContextProvider,
	)

	return ApiServerContext{}, nil
}
