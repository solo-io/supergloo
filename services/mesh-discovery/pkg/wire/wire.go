//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apps"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
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
		kubernetes_apps.DeploymentClientFactoryProvider,
		kubernetes_apps.ReplicaSetClientFactoryProvider,
		kubernetes_core.ServiceClientFactoryProvider,
		zephyr_discovery.MeshServiceClientFactoryProvider,
		zephyr_discovery.MeshWorkloadClientFactoryProvider,
		controllers.NewDeploymentEventWatcherFactory,
		controllers.NewPodEventWatcherFactory,
		controllers.NewServiceEventWatcherFactory,
		controllers.NewMeshWorkloadEventWatcherFactory,
		zephyr_discovery.MeshClientFactoryProvider,
		kubernetes_core.ConfigMapClientFactoryProvider,
		mesh_istio.WireProviderSet,
		mesh_consul.WireProviderSet,
		mesh_linkerd.WireProviderSet,
		DiscoveryContextProvider,
	)

	return DiscoveryContext{}, nil
}
