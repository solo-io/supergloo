package wire

import (
	"github.com/google/wire"
	istio_networking_providers "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/providers"
	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	linkerd_networking_providers "github.com/solo-io/external-apis/pkg/api/linkerd/linkerd.io/v1alpha2/providers"
	smi_networking_providers "github.com/solo-io/external-apis/pkg/api/smi/split.smi-spec.io/v1alpha1/providers"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/routing/traffic-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/routing/traffic-policy-translator/istio-translator"
	linkerd_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/routing/traffic-policy-translator/linkerd-translator"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/routing/traffic-policy-translator/preprocess"
)

var (
	TrafficPolicyProviderSet = wire.NewSet(
		smh_discovery.NewMeshServiceClient,
		smh_networking.NewTrafficPolicyClient,
		k8s_core_providers.ServiceClientFactoryProvider,
		istio_networking_providers.VirtualServiceClientFactoryProvider,
		istio_networking_providers.DestinationRuleClientFactoryProvider,
		linkerd_networking_providers.ServiceProfileClientFactoryProvider,
		smi_networking_providers.TrafficSplitClientFactoryProvider,
		istio_translator.NewIstioTrafficPolicyTranslator,
		linkerd_translator.NewLinkerdTrafficPolicyTranslator,
		TrafficPolicyMeshTranslatorsProvider,
		LocalTrafficPolicyEventWatcherProvider,
		traffic_policy_translator.NewTrafficPolicyTranslatorLoop,
		preprocess.NewTrafficPolicyPreprocessor,
		preprocess.NewTrafficPolicyMerger,
		selection.NewBaseResourceSelector,
		selection.NewResourceSelector,
		preprocess.NewTrafficPolicyValidator,
	)
)

func LocalTrafficPolicyEventWatcherProvider(mgr mc_manager.AsyncManager) smh_networking_controller.TrafficPolicyEventWatcher {
	return smh_networking_controller.NewTrafficPolicyEventWatcher("management-plane-traffic-policy-event-watcher", mgr.Manager())
}

func LocalMeshServiceEventWatcherProvider(mgr mc_manager.AsyncManager) smh_discovery_controller.MeshServiceEventWatcher {
	return smh_discovery_controller.NewMeshServiceEventWatcher("management-plane-mesh-service-event-watcher", mgr.Manager())
}

func LocalMeshWorkloadEventWatcherProvider(mgr mc_manager.AsyncManager) smh_discovery_controller.MeshWorkloadEventWatcher {
	return smh_discovery_controller.NewMeshWorkloadEventWatcher("management-plane-mesh-workload-event-watcher", mgr.Manager())
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
