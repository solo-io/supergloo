package wire

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload"
	mesh_consul "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh/linkerd"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws"
	event_watcher_factories "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/event-watcher-factories"
)

// just used to package everything up for wire
type DiscoveryContext struct {
	MultiClusterDeps      multicluster.MultiClusterDependencies
	ClientFactories       ClientFactories
	EventWatcherFactories EventWatcherFactories
	MeshDiscovery         MeshDiscovery
	RestAPIReconcilers    RestAPIReconcilers
}

type ClientFactories struct {
	ReplicaSetClientFactory   k8s_apps.ReplicaSetClientFactory
	DeploymentClientFactory   k8s_apps.DeploymentClientFactory
	OwnerFetcherClientFactory mesh_workload.OwnerFetcherFactory
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
	IstioMeshScanner         mesh_istio.IstioMeshScanner
	ConsulConnectMeshScanner mesh_consul.ConsulConnectMeshScanner
	LinkerdMeshScanner       mesh_linkerd.LinkerdMeshScanner
}

type RestAPIReconcilers struct {
	AppMeshAPIReconcilerFactory aws.AppMeshDiscoveryReconcilerFactory
}

func DiscoveryContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
	istioMeshScanner mesh_istio.IstioMeshScanner,
	consulConnectMeshScanner mesh_consul.ConsulConnectMeshScanner,
	linkerdMeshScanner mesh_linkerd.LinkerdMeshScanner,
	replicaSetClientFactory k8s_apps.ReplicaSetClientFactory,
	deploymentClientFactory k8s_apps.DeploymentClientFactory,
	ownerFetcherClientFactory mesh_workload.OwnerFetcherFactory,
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
	appMeshAPIReconcilerFactory aws.AppMeshDiscoveryReconcilerFactory,
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
			IstioMeshScanner:         istioMeshScanner,
			ConsulConnectMeshScanner: consulConnectMeshScanner,
			LinkerdMeshScanner:       linkerdMeshScanner,
		},
		RestAPIReconcilers: RestAPIReconcilers{
			AppMeshAPIReconcilerFactory: appMeshAPIReconcilerFactory,
		},
	}
}
