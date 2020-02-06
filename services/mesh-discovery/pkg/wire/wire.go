//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	multicluster_wire "github.com/solo-io/mesh-projects/services/common/multicluster/wire"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/consul"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/istio"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/linkerd"
)

func InitializeMeshDiscovery(ctx context.Context) (MeshDiscoveryContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		docker.NewImageNameParser,
		istio.WireProviderSet,
		consul.WireProviderSet,
		linkerd.WireProviderSet,
		MeshDiscoveryContextProvider,
	)

	return MeshDiscoveryContext{}, nil
}
