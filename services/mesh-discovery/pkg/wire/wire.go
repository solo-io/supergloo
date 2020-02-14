//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	multicluster_wire "github.com/solo-io/mesh-projects/services/common/multicluster/wire"
	mesh_consul "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/consul"
	mesh_istio "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/istio"
	mesh_linkerd "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/linkerd"
)

func InitializeDiscovery(ctx context.Context) (DiscoveryContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		docker.NewImageNameParser,
		mesh_istio.WireProviderSet,
		mesh_consul.WireProviderSet,
		mesh_linkerd.WireProviderSet,
		DiscoveryContextProvider,
	)

	return DiscoveryContext{}, nil
}
