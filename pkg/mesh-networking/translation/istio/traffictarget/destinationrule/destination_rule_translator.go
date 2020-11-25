package destinationrule

import (
	"context"
	"reflect"

	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/tls"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/selectorutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/snapshotutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"istio.io/istio/pkg/config/host"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

//go:generate mockgen -source ./destination_rule_translator.go -destination mocks/destination_rule_translator.go

// the DestinationRule translator translates a TrafficTarget into a DestinationRule.
type Translator interface {
	// Translate translates the appropriate DestinationRule for the given TrafficTarget.
	// returns nil if no DestinationRule is required for the TrafficTarget (i.e. if no DestinationRule features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot TrafficTargetSet contains the given TrafficTarget.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
		reporter reporting.Reporter,
	) *networkingv1alpha3.DestinationRule
}

type translator struct {
	clusterDomains   hostutils.ClusterDomainRegistry
	decoratorFactory decorators.Factory
	trafficTargets   discoveryv1alpha2sets.TrafficTargetSet
	failoverServices v1alpha2sets.FailoverServiceSet
}

func NewTranslator(
	clusterDomains hostutils.ClusterDomainRegistry,
	decoratorFactory decorators.Factory,
	trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	failoverServices v1alpha2sets.FailoverServiceSet,
) Translator {
	return &translator{
		clusterDomains:   clusterDomains,
		decoratorFactory: decoratorFactory,
		trafficTargets:   trafficTargets,
		failoverServices: failoverServices,
	}
}

// translate the appropriate DestinationRule for the given TrafficTarget.
// returns nil if no DestinationRule is required for the TrafficTarget (i.e. if no DestinationRule features are required, such as subsets).
// The input snapshot TrafficTargetSet contains n the
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
	reporter reporting.Reporter,
) *networkingv1alpha3.DestinationRule {
	settings, err := snapshotutils.GetSingletonSettings(ctx, in)
	if err != nil {
		// should be caught earlier in the reconciliation loop
		return nil
	}

	kubeService := trafficTarget.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	sourceClusterName := kubeService.Ref.ClusterName
	if sourceMeshInstallation != nil {
		sourceClusterName = sourceMeshInstallation.Cluster
	}

	destinationRule, err := t.initializeDestinationRule(trafficTarget, settings.Spec.Mtls, sourceMeshInstallation)
	if err != nil {
		contextutils.LoggerFrom(ctx).Error(err)
		return nil
	}

	// register the owners of the destinationrule fields
	destinationRuleFields := fieldutils.NewOwnershipRegistry()
	drDecorators := t.decoratorFactory.MakeDecorators(decorators.Parameters{
		ClusterDomains: t.clusterDomains,
		Snapshot:       in,
	})

	// Apply decorators which map a single applicable TrafficPolicy to a field on the DestinationRule.
	for _, policy := range trafficTarget.Status.AppliedTrafficPolicies {

		// Don't translate the trafficPolicy if the sourceClusterName is not selected by the SourceSelectors
		if !selectorutils.WorkloadSelectorContainsCluster(policy.Spec.SourceSelector, sourceClusterName) {
			continue
		}

		registerField := registerFieldFunc(destinationRuleFields, destinationRule, policy.Ref)
		for _, decorator := range drDecorators {

			if destinationRuleDecorator, ok := decorator.(decorators.TrafficPolicyDestinationRuleDecorator); ok {
				if err := destinationRuleDecorator.ApplyTrafficPolicyToDestinationRule(
					policy,
					trafficTarget,
					&destinationRule.Spec,
					registerField,
				); err != nil {
					reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, policy.Ref, eris.Wrapf(err, "%v", decorator.DecoratorName()))
				}
			}
		}
	}

	// TODO need a more robust implementation of determining whether a DestinationRule has any effect
	if len(destinationRule.Spec.Subsets) == 0 &&
		destinationRule.Spec.GetTrafficPolicy().GetTls().GetMode() == networkingv1alpha3spec.ClientTLSSettings_DISABLE &&
		destinationRule.Spec.GetTrafficPolicy().GetOutlierDetection() == nil {
		// no need to create this DestinationRule as it has no effect
		return nil
	}

	// detect and report error on intersecting config if enabled in settings
	if errs := conflictsWithUserDestinationRule(
		settings.Spec.GetNetworking().GetDisallowIntersectingConfig(),
		in.DestinationRules(),
		destinationRule,
	); len(errs) > 0 {
		for _, err := range errs {
			for _, policy := range trafficTarget.Status.AppliedTrafficPolicies {
				reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, policy.Ref, err)
			}
		}
		return nil
	}

	return destinationRule
}

// construct the callback for registering fields in the virtual service
func registerFieldFunc(
	destinationRuleFields fieldutils.FieldOwnershipRegistry,
	destinationRule *networkingv1alpha3.DestinationRule,
	policy ezkube.ResourceId,
) decorators.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.Equals(fieldVal, val) {
			return nil
		}
		if err := destinationRuleFields.RegisterFieldOwnership(
			destinationRule,
			fieldPtr,
			[]ezkube.ResourceId{policy},
			&v1alpha2.TrafficPolicy{},
			0, //TODO(ilackarms): priority
		); err != nil {
			return err
		}
		return nil
	}
}

func (t *translator) initializeDestinationRule(
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	mtlsDefault *v1alpha2.TrafficPolicySpec_MTLS,
	sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
) (*networkingv1alpha3.DestinationRule, error) {
	var meta metav1.ObjectMeta
	if sourceMeshInstallation != nil {
		meta = metautils.FederatedObjectMeta(
			trafficTarget.Spec.GetKubeService().Ref,
			sourceMeshInstallation,
			trafficTarget.Annotations,
		)
	} else {
		meta = metautils.TranslatedObjectMeta(
			trafficTarget.Spec.GetKubeService().Ref,
			trafficTarget.Annotations,
		)
	}
	hostname := t.clusterDomains.GetDestinationServiceFQDN(meta.ClusterName, trafficTarget.Spec.GetKubeService().Ref)

	destinationRule := &networkingv1alpha3.DestinationRule{
		ObjectMeta: meta,
		Spec: networkingv1alpha3spec.DestinationRule{
			Host:          hostname,
			TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{},
			Subsets: trafficshift.MakeDestinationRuleSubsetsForTrafficTarget(
				trafficTarget,
				t.trafficTargets,
				t.failoverServices,
				sourceMeshInstallation.GetCluster(),
			),
		},
	}

	// Initialize Istio TLS mode with default declared in Settings
	istioTlsMode, err := tls.MapIstioTlsMode(mtlsDefault.GetIstio().GetTlsMode())
	if err != nil {
		return nil, err
	}
	destinationRule.Spec.TrafficPolicy.Tls = &networkingv1alpha3spec.ClientTLSSettings{
		Mode: istioTlsMode,
	}

	return destinationRule, nil
}

// Return errors for each user-supplied VirtualService that applies to the same hostname as the translated VirtualService
func conflictsWithUserDestinationRule(
	disallowIntersectingConfig bool,
	destinationRules v1alpha3sets.DestinationRuleSet,
	translatedDestinationRule *networkingv1alpha3.DestinationRule,
) []error {
	if !disallowIntersectingConfig {
		return nil
	}

	// For each user DR, check whether any hosts match any hosts from translated DR
	var errs []error

	destinationRules.List(func(dr *networkingv1alpha3.DestinationRule) (_ bool) {
		// ignore translated DR
		if metautils.IsTranslated(dr) {
			return
		}
		// check if common hostnames exist
		userHostnames := host.NewNames([]string{dr.Spec.Host})
		translatedHostnames := host.NewNames([]string{translatedDestinationRule.Spec.Host})
		commonHostname := userHostnames.Intersection(translatedHostnames)

		if len(commonHostname) > 0 {
			errs = append(
				errs,
				eris.Errorf("Unable to translate AppliedTrafficPolicies to DestinationRule, applies to host %s that is already configured by the existing DestinationRule %s", commonHostname[0], sets.Key(dr)),
			)
		}
		return
	})

	return errs
}
