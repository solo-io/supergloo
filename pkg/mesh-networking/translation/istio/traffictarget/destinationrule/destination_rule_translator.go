package destinationrule

import (
	"context"
	"reflect"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/tls"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/snapshotutils"

	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
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
	reporter reporting.Reporter,
) *networkingv1alpha3.DestinationRule {
	kubeService := trafficTarget.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	settings, err := snapshotutils.GetSingletonSettings(ctx, in)
	if err != nil {
		return nil
	}

	destinationRule, err := t.initializeDestinationRule(trafficTarget, settings.Spec.Mtls)
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
		destinationRule.Spec.GetTrafficPolicy().GetTls().GetMode() == networkingv1alpha3spec.ClientTLSSettings_DISABLE {
		// no need to create this DestinationRule as it has no effect
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
) (*networkingv1alpha3.DestinationRule, error) {
	meta := metautils.TranslatedObjectMeta(
		trafficTarget.Spec.GetKubeService().Ref,
		trafficTarget.Annotations,
	)
	hostname := t.clusterDomains.GetServiceLocalFQDN(trafficTarget.Spec.GetKubeService().Ref)

	destinationRule := &networkingv1alpha3.DestinationRule{
		ObjectMeta: meta,
		Spec: networkingv1alpha3spec.DestinationRule{
			Host:          hostname,
			TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{},
			Subsets: trafficshift.MakeDestinationRuleSubsetsForTrafficTarget(
				trafficTarget,
				t.trafficTargets,
				t.failoverServices,
			),
		},
	}

	// Initialize Istio TLS mode with default declared in Settings
	istioTlsMode, err := tls.MapIstioTlsMode(mtlsDefault.Istio.TlsMode)
	if err != nil {
		return nil, err
	}
	destinationRule.Spec.TrafficPolicy.Tls = &networkingv1alpha3spec.ClientTLSSettings{
		Mode: istioTlsMode,
	}

	return destinationRule, nil
}
