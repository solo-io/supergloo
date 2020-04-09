package wire

import (
	"github.com/google/wire"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/server"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
)

// just used to package everything up for wire
type ApiServerContext struct {
	MultiClusterDeps       multicluster.MultiClusterDependencies
	ManagementPlaneClients ManagementPlaneClients
	Server                 server.GrpcServer
}

var managementPlaneClientsSet = wire.NewSet(
	zephyr_discovery.NewMeshServiceClient,
	zephyr_discovery.NewMeshWorkloadClient,
	zephyr_discovery.NewMeshClient,
	zephyr_discovery.NewControllerRuntimeKubernetesClusterClient,
)

type ManagementPlaneClients struct {
	MeshServiceClient       zephyr_discovery.MeshServiceClient
	MeshWorkloadClient      zephyr_discovery.MeshWorkloadClient
	MeshClient              zephyr_discovery.MeshClient
	KubernetesClusterClient zephyr_discovery.KubernetesClusterClient
}

func ApiServerContextProvider(
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshClient zephyr_discovery.MeshClient,
	kubernetesClusterClient zephyr_discovery.KubernetesClusterClient,
	grpcServer server.GrpcServer,
	multiClusterDeps multicluster.MultiClusterDependencies,
) ApiServerContext {

	return ApiServerContext{
		ManagementPlaneClients: ManagementPlaneClients{
			MeshServiceClient:       meshServiceClient,
			MeshWorkloadClient:      meshWorkloadClient,
			MeshClient:              meshClient,
			KubernetesClusterClient: kubernetesClusterClient,
		},
		Server:           grpcServer,
		MultiClusterDeps: multiClusterDeps,
	}
}
