package wire

import (
	"github.com/google/wire"
	discovery_controller "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	networking_controller "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	istio_networking "github.com/solo-io/mesh-projects/pkg/clients/istio/networking"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	networking_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/selector"
	traffic_policy_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator"
	istio_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
)

var (
	TrafficPolicyProviderSet = wire.NewSet(
		discovery_core.NewMeshServiceClient,
		networking_core.NewTrafficPolicyClient,
		kubernetes_core.ServiceClientFactoryProvider,
		istio_networking.VirtualServiceClientFactoryProvider,
		istio_networking.DestinationRuleClientFactoryProvider,
		istio_translator.NewIstioTrafficPolicyTranslator,
		TrafficPolicyMeshTranslatorsProvider,
		LocalTrafficPolicyControllerProvider,
		LocalMeshServiceControllerProvider,
		traffic_policy_translator.NewTrafficPolicyTranslatorLoop,
		preprocess.NewTrafficPolicyPreprocessor,
		preprocess.NewTrafficPolicyMerger,
		selector.NewMeshServiceSelector,
		preprocess.NewTrafficPolicyValidator,
	)
)

func LocalTrafficPolicyControllerProvider(mgr mc_manager.AsyncManager) (networking_controller.TrafficPolicyController, error) {
	return networking_controller.NewTrafficPolicyController("management-plane-traffic-policy-controller", mgr.Manager())
}

func LocalMeshServiceControllerProvider(mgr mc_manager.AsyncManager) (discovery_controller.MeshServiceController, error) {
	return discovery_controller.NewMeshServiceController("management-plane-mesh-service-controller", mgr.Manager())
}

func TrafficPolicyMeshTranslatorsProvider(
	istioTranslator istio_translator.IstioTranslator,
) []traffic_policy_translator.TrafficPolicyMeshTranslator {
	return []traffic_policy_translator.TrafficPolicyMeshTranslator{istioTranslator}
}
