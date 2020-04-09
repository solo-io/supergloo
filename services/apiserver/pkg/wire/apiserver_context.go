package wire

import (
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
)

// just used to package everything up for wire
type ApiServerContext struct {
	MultiClusterDeps multicluster.MultiClusterDependencies
}

func ApiServerContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
) ApiServerContext {

	return ApiServerContext{
		MultiClusterDeps: multiClusterDeps,
	}
}
