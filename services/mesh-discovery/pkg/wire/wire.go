//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apps"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	multicluster_wire "github.com/solo-io/service-mesh-hub/services/common/multicluster/wire"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload"
	mesh_consul "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/linkerd"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/multicluster/controllers"
)

func InitializeDiscovery(ctx context.Context) (DiscoveryContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		docker.NewImageNameParser,
		mesh_workload.OwnerFetcherFactoryProvider,
		kubernetes_apps.ControllerRuntimeDeploymentClientFactoryProvider,
		kubernetes_apps.ReplicaSetClientFactoryProvider,
		kubernetes_core.ServiceClientFactoryProvider,
		discovery_core.MeshServiceClientFactoryProvider,
		discovery_core.MeshWorkloadClientFactoryProvider,
		controllers.NewDeploymentControllerFactory,
		controllers.NewPodControllerFactory,
		controllers.NewServiceControllerFactory,
		controllers.NewMeshWorkloadControllerFactory,
		controllers.NewMeshControllerFactory,
		discovery_core.NewMeshClientFactoryProvider,
		kubernetes_core.ConfigMapClientFactoryProvider,
		kubernetes_core.NewPodClientFactory,
		mesh_istio.WireProviderSet,
		mesh_consul.WireProviderSet,
		mesh_linkerd.WireProviderSet,
		DiscoveryContextProvider,
	)

	return DiscoveryContext{}, nil
}
