package wire

import (
	"github.com/google/wire"
	discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/clients/istio/networking"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	networking_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
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
		selector.NewResourceSelector,
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
