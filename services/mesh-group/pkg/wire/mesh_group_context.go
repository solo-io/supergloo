package wire

import (
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
)

// just used to package everything up for wire
type MeshGroupContext struct {
	MultiClusterDeps      multicluster.MultiClusterDependencies
	MeshGroupEventHandler controller.MeshGroupEventHandler
}

func MeshGroupContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
	meshGroupEventHandler controller.MeshGroupEventHandler,
) MeshGroupContext {
	return MeshGroupContext{
		MultiClusterDeps:      multiClusterDeps,
		MeshGroupEventHandler: meshGroupEventHandler,
	}
}
