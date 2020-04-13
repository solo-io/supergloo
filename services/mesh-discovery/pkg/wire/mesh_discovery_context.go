package wire

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apps"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload"
	mesh_consul "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/consul"
	mesh_istio "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/istio"
	mesh_linkerd "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/linkerd"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/multicluster/controllers"
)

// just used to package everything up for wire
type DiscoveryContext struct {
	MultiClusterDeps      multicluster.MultiClusterDependencies
	ClientFactories       ClientFactories
	EventWatcherFactories EventWatcherFactories
	MeshDiscovery         MeshDiscovery
}

type ClientFactories struct {
	ReplicaSetClientFactory   kubernetes_apps.ReplicaSetClientFactory
	DeploymentClientFactory   kubernetes_apps.DeploymentClientFactory
	OwnerFetcherClientFactory mesh_workload.OwnerFetcherFactory
	ServiceClientFactory      kubernetes_core.ServiceClientFactory
	MeshServiceClientFactory  zephyr_discovery.MeshServiceClientFactory
	MeshWorkloadClientFactory zephyr_discovery.MeshWorkloadClientFactory
	MeshClientFactory         zephyr_discovery.MeshClientFactory
}

type EventWatcherFactories struct {
	DeploymentEventWatcherFactory   controllers.DeploymentEventWatcherFactory
	PodEventWatcherFactory          controllers.PodEventWatcherFactory
	ServiceEventWatcherFactory      controllers.ServiceControllerFactory
	MeshWorkloadEventWatcherFactory controllers.MeshWorkloadEventWatcherFactory
}

type MeshDiscovery struct {
	IstioMeshScanner         mesh_istio.IstioMeshScanner
	ConsulConnectMeshScanner mesh_consul.ConsulConnectMeshScanner
	LinkerdMeshScanner       mesh_linkerd.LinkerdMeshScanner
}

func DiscoveryContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
	istioMeshScanner mesh_istio.IstioMeshScanner,
	consulConnectMeshScanner mesh_consul.ConsulConnectMeshScanner,
	linkerdMeshScanner mesh_linkerd.LinkerdMeshScanner,
	replicaSetClientFactory kubernetes_apps.ReplicaSetClientFactory,
	deploymentClientFactory kubernetes_apps.DeploymentClientFactory,
	ownerFetcherClientFactory mesh_workload.OwnerFetcherFactory,
	serviceClientFactory kubernetes_core.ServiceClientFactory,
	meshServiceClientFactory zephyr_discovery.MeshServiceClientFactory,
	meshWorkloadClientFactory zephyr_discovery.MeshWorkloadClientFactory,
	podEventWatcherFactory controllers.PodEventWatcherFactory,
	serviceEventWatcherFactory controllers.ServiceControllerFactory,
	meshWorkloadControllerFactory controllers.MeshWorkloadEventWatcherFactory,
	deploymentEventWatcherFactory controllers.DeploymentEventWatcherFactory,
	meshClientFactory zephyr_discovery.MeshClientFactory,
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
		},
		EventWatcherFactories: EventWatcherFactories{
			DeploymentEventWatcherFactory:   deploymentEventWatcherFactory,
			PodEventWatcherFactory:          podEventWatcherFactory,
			ServiceEventWatcherFactory:      serviceEventWatcherFactory,
			MeshWorkloadEventWatcherFactory: meshWorkloadControllerFactory,
		},
		MeshDiscovery: MeshDiscovery{
			IstioMeshScanner:         istioMeshScanner,
			ConsulConnectMeshScanner: consulConnectMeshScanner,
			LinkerdMeshScanner:       linkerdMeshScanner,
		},
	}
}
