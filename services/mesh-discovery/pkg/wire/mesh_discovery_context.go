package wire

import (
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mesh_consul "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/consul"
	mesh_istio "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/istio"
	mesh_linkerd "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh/linkerd"
)

// just used to package everything up for wire
type DiscoveryContext struct {
	ImageParser      docker.ImageNameParser
	MultiClusterDeps multicluster.MultiClusterDependencies
	// Mesh discovery
	IstioMeshScanner         mesh_istio.IstioMeshScanner
	ConsulConnectMeshScanner mesh_consul.ConsulConnectMeshScanner
	LinkerdMeshScanner       mesh_linkerd.LinkerdMeshScanner
}

func DiscoveryContextProvider(
	imageParser docker.ImageNameParser,
	multiClusterDeps multicluster.MultiClusterDependencies,
	istioMeshScanner mesh_istio.IstioMeshScanner,
	consulConnectMeshScanner mesh_consul.ConsulConnectMeshScanner,
	linkerdMeshScanner mesh_linkerd.LinkerdMeshScanner,
) DiscoveryContext {
	return DiscoveryContext{
		ImageParser:              imageParser,
		MultiClusterDeps:         multiClusterDeps,
		IstioMeshScanner:         istioMeshScanner,
		ConsulConnectMeshScanner: consulConnectMeshScanner,
		LinkerdMeshScanner:       linkerdMeshScanner,
	}
}
