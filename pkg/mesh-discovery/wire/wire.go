//+build wireinject

package wire

import (
	"context"

	k8s_apps_providers "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/providers"
	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	v1alpha1_providers "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/providers"
	smh_discovery_providers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/providers"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/selection"
	settings_clients "github.com/solo-io/service-mesh-hub/pkg/common/aws/settings"
	multicluster_wire "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/wire"
	"github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/common/filesystem/files"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	event_watcher_factories "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/event-watcher-factories"
	appmesh_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s/appmesh"
	mesh_consul "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/linkerd"
)

func InitializeDiscovery(ctx context.Context) (DiscoveryContext, error) {
	wire.Build(
		multicluster_wire.MulticlusterProviderSet,
		docker.NewImageNameParser,
		k8s_apps_providers.DeploymentClientFactoryProvider,
		k8s_apps_providers.ReplicaSetClientFactoryProvider,
		k8s_core_providers.ServiceClientFactoryProvider,
		k8s_core_providers.PodClientFactoryProvider,
		smh_discovery_providers.MeshServiceClientFactoryProvider,
		smh_discovery_providers.MeshWorkloadClientFactoryProvider,
		event_watcher_factories.NewDeploymentEventWatcherFactory,
		event_watcher_factories.NewPodEventWatcherFactory,
		event_watcher_factories.NewServiceEventWatcherFactory,
		event_watcher_factories.NewMeshWorkloadEventWatcherFactory,
		event_watcher_factories.NewMeshEventWatcherFactory,
		smh_discovery_providers.MeshClientFactoryProvider,
		k8s_core_providers.ConfigMapClientFactoryProvider,
		mesh_istio.WireProviderSet,
		mesh_consul.WireProviderSet,
		mesh_linkerd.WireProviderSet,
		DiscoveryContextProvider,
		AwsSet,
		kubeconfig.NewConverter,
		files.NewDefaultFileReader,
		appmesh_tenancy.AppMeshTenancyScannerFactoryProvider,
		ClusterRegistrationSet,
		ComputeTargetCredentialsHandlersProvider,
		ClusterRegistrationClientProvider,
		v1alpha1_providers.SettingsClientProvider,
		settings_clients.NewAwsSettingsHelperClient,
		selection.NewAwsSelector,
		clients.STSClientFactoryProvider,
	)

	return DiscoveryContext{}, nil
}
