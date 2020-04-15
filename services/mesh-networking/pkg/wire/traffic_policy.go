package wire

import (
	"github.com/google/wire"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	linkerd_networking "github.com/solo-io/service-mesh-hub/pkg/api/linkerd/v1alpha2"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	smi_networking "github.com/solo-io/service-mesh-hub/pkg/api/smi/split/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	linkerd_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/linkerd-translator"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
)

var (
	TrafficPolicyProviderSet = wire.NewSet(
		zephyr_discovery.NewMeshServiceClient,
		zephyr_networking.NewTrafficPolicyClient,
		kubernetes_core.ServiceClientFactoryProvider,
		istio_networking.VirtualServiceClientFactoryProvider,
		istio_networking.DestinationRuleClientFactoryProvider,
		linkerd_networking.ServiceProfileClientFactoryProvider,
		smi_networking.TrafficSplitClientFactoryProvider,
		istio_translator.NewIstioTrafficPolicyTranslator,
		linkerd_translator.NewLinkerdTrafficPolicyTranslator,
		TrafficPolicyMeshTranslatorsProvider,
		LocalTrafficPolicyEventWatcherProvider,
		traffic_policy_translator.NewTrafficPolicyTranslatorLoop,
		preprocess.NewTrafficPolicyPreprocessor,
		preprocess.NewTrafficPolicyMerger,
		selector.NewResourceSelector,
		preprocess.NewTrafficPolicyValidator,
	)
)

func LocalTrafficPolicyEventWatcherProvider(mgr mc_manager.AsyncManager) networking_controller.TrafficPolicyEventWatcher {
	return networking_controller.NewTrafficPolicyEventWatcher("management-plane-traffic-policy-event-watcher", mgr.Manager())
}

func LocalMeshServiceEventWatcherProvider(mgr mc_manager.AsyncManager) discovery_controller.MeshServiceEventWatcher {
	return discovery_controller.NewMeshServiceEventWatcher("management-plane-mesh-service-event-watcher", mgr.Manager())
}

func LocalMeshWorkloadEventWatcherProvider(mgr mc_manager.AsyncManager) discovery_controller.MeshWorkloadEventWatcher {
	return discovery_controller.NewMeshWorkloadEventWatcher("management-plane-mesh-workload-event-watcher", mgr.Manager())
}

func TrafficPolicyMeshTranslatorsProvider(
	istioTranslator istio_translator.IstioTranslator,
	linkerdTranslator linkerd_translator.LinkerdTranslator,
) []traffic_policy_translator.TrafficPolicyMeshTranslator {
	return []traffic_policy_translator.TrafficPolicyMeshTranslator{
		istioTranslator,
		linkerdTranslator,
	}
}
