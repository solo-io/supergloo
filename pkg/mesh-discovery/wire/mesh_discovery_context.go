package wire

import (
	k8s_apps_providers "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/providers"
	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	smh_discovery_providers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/providers"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	event_watcher_factories "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/event-watcher-factories"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	mesh_consul "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/linkerd"
)

// just used to package everything up for wire
type DiscoveryContext struct {
	MultiClusterDeps      mc_manager.MultiClusterDependencies
	ClientFactories       ClientFactories
	EventWatcherFactories EventWatcherFactories
	MeshDiscovery         MeshDiscovery
	ClusterTenancy        ClusterTenancy
}

type ClientFactories struct {
	ReplicaSetClientFactory   k8s_apps_providers.ReplicaSetClientFactory
	DeploymentClientFactory   k8s_apps_providers.DeploymentClientFactory
	OwnerFetcherClientFactory meshworkload_discovery.OwnerFetcherFactory
	ServiceClientFactory      k8s_core_providers.ServiceClientFactory
	MeshServiceClientFactory  smh_discovery_providers.MeshServiceClientFactory
	MeshWorkloadClientFactory smh_discovery_providers.MeshWorkloadClientFactory
	MeshClientFactory         smh_discovery_providers.MeshClientFactory
	PodClientFactory          k8s_core_providers.PodClientFactory
}

type EventWatcherFactories struct {
	DeploymentEventWatcherFactory   event_watcher_factories.DeploymentEventWatcherFactory
	PodEventWatcherFactory          event_watcher_factories.PodEventWatcherFactory
	ServiceEventWatcherFactory      event_watcher_factories.ServiceEventWatcherFactory
	MeshWorkloadEventWatcherFactory event_watcher_factories.MeshWorkloadEventWatcherFactory
	MeshControllerFactory           event_watcher_factories.MeshEventWatcherFactory
}

type MeshDiscovery struct {
	IstioMeshScanner              mesh_istio.IstioMeshScanner
	ConsulConnectMeshScanner      mesh_consul.ConsulConnectMeshScanner
	LinkerdMeshScanner            mesh_linkerd.LinkerdMeshScanner
	AppMeshWorkloadScannerFactory meshworkload_discovery.MeshWorkloadScannerFactory
}

type ClusterTenancy struct {
	AppMeshClusterTenancyScannerFactory k8s_tenancy.ClusterTenancyScannerFactory
}

func DiscoveryContextProvider(
	multiClusterDeps mc_manager.MultiClusterDependencies,
	istioMeshScanner mesh_istio.IstioMeshScanner,
	consulConnectMeshScanner mesh_consul.ConsulConnectMeshScanner,
	linkerdMeshScanner mesh_linkerd.LinkerdMeshScanner,
	replicaSetClientFactory k8s_apps_providers.ReplicaSetClientFactory,
	deploymentClientFactory k8s_apps_providers.DeploymentClientFactory,
	ownerFetcherClientFactory meshworkload_discovery.OwnerFetcherFactory,
	serviceClientFactory k8s_core_providers.ServiceClientFactory,
	meshServiceClientFactory smh_discovery_providers.MeshServiceClientFactory,
	meshWorkloadClientFactory smh_discovery_providers.MeshWorkloadClientFactory,
	podEventWatcherFactory event_watcher_factories.PodEventWatcherFactory,
	serviceEventWatcherFactory event_watcher_factories.ServiceEventWatcherFactory,
	meshWorkloadControllerFactory event_watcher_factories.MeshWorkloadEventWatcherFactory,
	deploymentEventWatcherFactory event_watcher_factories.DeploymentEventWatcherFactory,
	meshClientFactory smh_discovery_providers.MeshClientFactory,
	podClientFactory k8s_core_providers.PodClientFactory,
	meshControllerFactory event_watcher_factories.MeshEventWatcherFactory,
	appMeshWorkloadScannerFactory meshworkload_discovery.MeshWorkloadScannerFactory,
	appMeshClusterTenancyScannerFactory k8s_tenancy.ClusterTenancyScannerFactory,
) DiscoveryContext {

	return DiscoveryContext{
		MultiClusterDeps: multiClusterDeps,
		ClientFactories: ClientFactories{
			ReplicaSetClientFactory:   replicaSetClientFactory,
			DeploymentClientFactory:   deploymentClientFactory,
			OwnerFetcherClientFactory: ownerFetcherClientFactory,
			ServiceClientFactory:      serviceClientFactory,
			MeshServiceClientFactory:  meshServiceClientFactory,
			MeshWorkloadClientFactory: meshWorkloadClientFactory,
			MeshClientFactory:         meshClientFactory,
			PodClientFactory:          podClientFactory,
		},
		EventWatcherFactories: EventWatcherFactories{
			DeploymentEventWatcherFactory:   deploymentEventWatcherFactory,
			PodEventWatcherFactory:          podEventWatcherFactory,
			ServiceEventWatcherFactory:      serviceEventWatcherFactory,
			MeshWorkloadEventWatcherFactory: meshWorkloadControllerFactory,
			MeshControllerFactory:           meshControllerFactory,
		},
		MeshDiscovery: MeshDiscovery{
			IstioMeshScanner:              istioMeshScanner,
			ConsulConnectMeshScanner:      consulConnectMeshScanner,
			LinkerdMeshScanner:            linkerdMeshScanner,
			AppMeshWorkloadScannerFactory: appMeshWorkloadScannerFactory,
		},
		ClusterTenancy: ClusterTenancy{
			AppMeshClusterTenancyScannerFactory: appMeshClusterTenancyScannerFactory,
		},
	}
}
