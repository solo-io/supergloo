package wire

import (
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/consul"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/istio"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/linkerd"
)

// just used to package everything up for wire
type MeshDiscoveryContext struct {
	MultiClusterDeps        multicluster.MultiClusterDependencies
	IstioMeshFinder         istio.IstioMeshFinder
	ConsulConnectMeshFinder consul.ConsulConnectMeshFinder
	LinkerdMeshFinder       linkerd.LinkerdMeshFinder
}

func MeshDiscoveryContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
	istioMeshFinder istio.IstioMeshFinder,
	consulConnectMeshFinder consul.ConsulConnectMeshFinder,
	linkerdMeshFinder linkerd.LinkerdMeshFinder,
) MeshDiscoveryContext {
	return MeshDiscoveryContext{
		MultiClusterDeps:        multiClusterDeps,
		IstioMeshFinder:         istioMeshFinder,
		ConsulConnectMeshFinder: consulConnectMeshFinder,
		LinkerdMeshFinder:       linkerdMeshFinder,
	}
}
