package wire

import (
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

// just used to package everything up for wire
type MeshNetworkingContext struct {
	MultiClusterDeps             multicluster.MultiClusterDependencies
	MeshGroupEventHandler        controller.MeshGroupEventHandler
	MeshNetworkingClusterHandler mc_manager.AsyncManagerHandler
}

func MeshNetworkingContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
	meshGroupEventHandler controller.MeshGroupEventHandler,
	meshNetworkingClusterHandler mc_manager.AsyncManagerHandler,
) MeshNetworkingContext {
	return MeshNetworkingContext{
		MultiClusterDeps:             multiClusterDeps,
		MeshGroupEventHandler:        meshGroupEventHandler,
		MeshNetworkingClusterHandler: meshNetworkingClusterHandler,
	}
}
