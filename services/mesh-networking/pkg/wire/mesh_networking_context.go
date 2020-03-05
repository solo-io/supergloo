package wire

import (
	networking_controller "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	traffic_policy_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator"
)

// just used to package everything up for wire
type MeshNetworkingContext struct {
	MultiClusterDeps             multicluster.MultiClusterDependencies
	MeshGroupEventHandler        networking_controller.MeshGroupEventHandler
	MeshNetworkingClusterHandler mc_manager.AsyncManagerHandler
	TrafficPolicyTranslator      traffic_policy_translator.TrafficPolicyTranslator
}

func MeshNetworkingContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
	meshGroupEventHandler networking_controller.MeshGroupEventHandler,
	meshNetworkingClusterHandler mc_manager.AsyncManagerHandler,
	trafficPolicyTranslator traffic_policy_translator.TrafficPolicyTranslator,
) MeshNetworkingContext {
	return MeshNetworkingContext{
		MultiClusterDeps:             multiClusterDeps,
		MeshGroupEventHandler:        meshGroupEventHandler,
		MeshNetworkingClusterHandler: meshNetworkingClusterHandler,
		TrafficPolicyTranslator:      trafficPolicyTranslator,
	}
}
