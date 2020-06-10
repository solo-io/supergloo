package wire

import (
	"github.com/google/wire"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	linkerd_networking "github.com/solo-io/service-mesh-hub/pkg/api/linkerd/v1alpha2"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	smi_networking "github.com/solo-io/service-mesh-hub/pkg/api/smi/split/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	linkerd_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/linkerd-translator"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
)

var (
	TrafficPolicyProviderSet = wire.NewSet(
		smh_discovery.NewMeshServiceClient,
		smh_networking.NewTrafficPolicyClient,
		k8s_core.ServiceClientFactoryProvider,
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
