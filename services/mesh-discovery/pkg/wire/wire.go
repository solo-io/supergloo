//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/files"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	multicluster_wire "github.com/solo-io/service-mesh-hub/services/common/compute-target/wire"
	event_watcher_factories "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/event-watcher-factories"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	mesh_consul "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/k8s/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/k8s/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/k8s/linkerd"
)

func InitializeDiscovery(ctx context.Context) (DiscoveryContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		docker.NewImageNameParser,
		k8s.OwnerFetcherFactoryProvider,
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
		ComputeTargetCredentialsHandlersProvider,
		AwsSet,
		kube.NewConverter,
		files.NewDefaultFileReader,
	)

	return DiscoveryContext{}, nil
}
