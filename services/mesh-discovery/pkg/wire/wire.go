//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	multicluster_wire "github.com/solo-io/service-mesh-hub/services/common/multicluster/wire"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload"
	mesh_consul "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh/linkerd"
	event_watcher_factories "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/event-watcher-factories"
)

func InitializeDiscovery(ctx context.Context) (DiscoveryContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		docker.NewImageNameParser,
		mesh_workload.OwnerFetcherFactoryProvider,
		k8s_apps.DeploymentClientFactoryProvider,
		k8s_apps.ReplicaSetClientFactoryProvider,
		k8s_core.ServiceClientFactoryProvider,
		k8s_core.PodClientFactoryProvider,
		zephyr_discovery.MeshServiceClientFactoryProvider,
		zephyr_discovery.MeshWorkloadClientFactoryProvider,
		event_watcher_factories.NewDeploymentEventWatcherFactory,
		event_watcher_factories.NewPodEventWatcherFactory,
		event_watcher_factories.NewServiceEventWatcherFactory,
		event_watcher_factories.NewMeshWorkloadEventWatcherFactory,
		event_watcher_factories.NewMeshEventWatcherFactory,
		zephyr_discovery.MeshClientFactoryProvider,
		k8s_core.ConfigMapClientFactoryProvider,
		mesh_istio.WireProviderSet,
		mesh_consul.WireProviderSet,
		mesh_linkerd.WireProviderSet,
		DiscoveryContextProvider,
		RestAPIHandlersProvider,
		AwsSet,
	)

	return DiscoveryContext{}, nil
}
