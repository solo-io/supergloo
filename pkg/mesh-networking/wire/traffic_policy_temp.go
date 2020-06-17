package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/reconcile"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/aggregation"
	aggregation_framework "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/aggregation/framework"
	translation_framework "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/framework"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/framework/snapshot"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/framework/snapshot/reconcilers"
	mesh_translation_aggregate "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/translators/aggregate"
	istio_mesh_translation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/translators/istio"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/validation"
)

var (
	NewTrafficPolicyProviderSet = wire.NewSet(
		NewReconciler,
	)
)

func NewReconciler(
	trafficPolicyClient smh_networking.TrafficPolicyClient,
	meshServiceClient smh_discovery.MeshServiceClient,
	meshClient smh_discovery.MeshClient,
	dynamicClientGetter multicluster.DynamicClientGetter,
) reconcile.Reconciler {

	virtualServiceReconciler := reconcilers.NewVirtualServiceReconcilerBuilder()
	destinationRuleReconciler := reconcilers.NewDestinationRuleReconcilerBuilder()
	snapshotReconciler := snapshot.NewSnapshotReconciler(dynamicClientGetter, virtualServiceReconciler, destinationRuleReconciler)

	baseSelector := selection.NewBaseResourceSelector()
	validator := traffic_policy_validation.NewValidator(baseSelector)
	validationProcessor := traffic_policy_validation.NewValidationProcessor(validator)
	aggregator := traffic_policy_aggregation.NewAggregator(baseSelector)
	policyCollector := traffic_policy_aggregation.NewPolicyCollector(aggregator)
	inMemoryStatusMutator := traffic_policy_aggregation.NewInMemoryStatusMutator()

	istio := istio_mesh_translation.NewIstioTrafficPolicyTranslator(baseSelector)

	translationMap := mesh_translation_aggregate.NewMeshTranslatorFactory(istio)
	aggregationProcessor := aggregation_framework.NewAggregationProcessor(
		meshServiceClient,
		meshClient,
		policyCollector,
		translationMap.MeshTypeToTranslationValidator,
		inMemoryStatusMutator,
	)
	translationProcessor := translation_framework.NewTranslationProcessor(meshClient, translationMap.MeshTypeToAccumulator)

	return reconcile.NewReconciler(trafficPolicyClient, meshServiceClient, snapshotReconciler, validationProcessor, aggregationProcessor, translationProcessor)
}
