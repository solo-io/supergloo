package wire

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	event_watcher_factories "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/event-watcher-factories"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	mesh_consul "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/k8s/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/k8s/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/k8s/linkerd"
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
	ReplicaSetClientFactory   k8s_apps.ReplicaSetClientFactory
	DeploymentClientFactory   k8s_apps.DeploymentClientFactory
	OwnerFetcherClientFactory meshworkload_discovery.OwnerFetcherFactory
	ServiceClientFactory      k8s_core.ServiceClientFactory
	MeshServiceClientFactory  zephyr_discovery.MeshServiceClientFactory
	MeshWorkloadClientFactory zephyr_discovery.MeshWorkloadClientFactory
	MeshClientFactory         zephyr_discovery.MeshClientFactory
	PodClientFactory          k8s_core.PodClientFactory
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
	replicaSetClientFactory k8s_apps.ReplicaSetClientFactory,
	deploymentClientFactory k8s_apps.DeploymentClientFactory,
	ownerFetcherClientFactory meshworkload_discovery.OwnerFetcherFactory,
	serviceClientFactory k8s_core.ServiceClientFactory,
	meshServiceClientFactory zephyr_discovery.MeshServiceClientFactory,
	meshWorkloadClientFactory zephyr_discovery.MeshWorkloadClientFactory,
	podEventWatcherFactory event_watcher_factories.PodEventWatcherFactory,
	serviceEventWatcherFactory event_watcher_factories.ServiceEventWatcherFactory,
	meshWorkloadControllerFactory event_watcher_factories.MeshWorkloadEventWatcherFactory,
	deploymentEventWatcherFactory event_watcher_factories.DeploymentEventWatcherFactory,
	meshClientFactory zephyr_discovery.MeshClientFactory,
	podClientFactory k8s_core.PodClientFactory,
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
