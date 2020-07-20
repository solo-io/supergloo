package destinationrule

import (
	"reflect"

	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// the DestinationRule translator translates a MeshService into a DestinationRule.
type Translator interface {
	// Translate translates the appropriate DestinationRule for the given MeshService.
	// returns nil if no DestinationRule is required for the MeshService (i.e. if no DestinationRule features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot MeshServiceSet contains the given MeshService.
	Translate(
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		reporter reporting.Reporter,
	) *istiov1alpha3.DestinationRule
}

type translator struct {
	clusterDomains   hostutils.ClusterDomainRegistry
	decoratorFactory decorators.Factory
}

func NewTranslator(clusterDomains hostutils.ClusterDomainRegistry, decoratorFactory decorators.Factory) Translator {
	return &translator{clusterDomains: clusterDomains, decoratorFactory: decoratorFactory}
}

// translate the appropriate DestinationRUle for the given MeshService.
// returns nil if no DestinationRule is required for the MeshService (i.e. if no DestinationRule features are required, such as subsets).
// The input snapshot MeshServiceSet contains n the
func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha2.MeshService,
	reporter reporting.Reporter,
) *istiov1alpha3.DestinationRule {
	kubeService := meshService.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	destinationRule := t.initializeDestinationRule(meshService)
	// register the owners of the destinationrule fields
	destinationRuleFields := fieldutils.NewOwnershipRegistry()
	drDecorators := t.decoratorFactory.MakeDecorators(decorators.Parameters{
		ClusterDomains: t.clusterDomains,
		Snapshot:       in,
	})

	// Apply decorators which aggregate the entire set of applicable TrafficPolicies to a field on the DestinationRule.
	trafficPolicyResourceIds := t.trafficPolicyToResourceIds(meshService.Status.AppliedTrafficPolicies)
	registerField := registerFieldFunc(destinationRuleFields, destinationRule, trafficPolicyResourceIds)
	for _, decorator := range drDecorators {

		if aggregatingDestinationRuleDecorator, ok := decorator.(trafficpolicy.AggregatingDestinationRuleDecorator); ok {
			if err := aggregatingDestinationRuleDecorator.ApplyAllToDestinationRule(
				meshService.Status.AppliedTrafficPolicies,
				&destinationRule.Spec,
				registerField,
			); err != nil {
				for _, policyResourceId := range trafficPolicyResourceIds {
					reporter.ReportTrafficPolicy(meshService, policyResourceId, eris.Wrapf(err, "%v", decorator.DecoratorName()))
				}
			}
		}
	}

	// Apply decorators which map a single applicable TrafficPolicy to a field on the DestinationRule.
	for _, policy := range meshService.Status.AppliedTrafficPolicies {
		registerField := registerFieldFunc(destinationRuleFields, destinationRule, []ezkube.ResourceId{policy.Ref})
		for _, decorator := range drDecorators {

			if destinationRuleDecorator, ok := decorator.(trafficpolicy.DestinationRuleDecorator); ok {
				if err := destinationRuleDecorator.ApplyToDestinationRule(
					policy,
					meshService,
					&destinationRule.Spec,
					registerField,
				); err != nil {
					reporter.ReportTrafficPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", decorator.DecoratorName()))
				}
			}
		}
	}

	if len(destinationRule.Spec.Subsets) == 0 && destinationRule.Spec.TrafficPolicy == nil {
		// no need to create this DestinationRule as it has no effect
		return nil
	}

	return destinationRule
}

// construct the callback for registering fields in the virtual service
func registerFieldFunc(
	destinationRuleFields fieldutils.FieldOwnershipRegistry,
	destinationRule *istiov1alpha3.DestinationRule,
	policyRefs []ezkube.ResourceId,
) decorators.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.Equals(fieldVal, val) {
			return nil
		}
		if err := destinationRuleFields.RegisterFieldOwnership(
			destinationRule,
			fieldPtr,
			policyRefs,
			&v1alpha2.TrafficPolicy{},
			0, //TODO(ilackarms): priority
		); err != nil {
			return err
		}
		return nil
	}
}

func (t *translator) initializeDestinationRule(meshService *discoveryv1alpha2.MeshService) *istiov1alpha3.DestinationRule {
	meta := metautils.TranslatedObjectMeta(
		meshService.Spec.GetKubeService().Ref,
		meshService.Annotations,
	)
	hostname := t.clusterDomains.GetServiceLocalFQDN(meshService.Spec.GetKubeService().Ref)

	return &istiov1alpha3.DestinationRule{
		ObjectMeta: meta,
		Spec: istiov1alpha3spec.DestinationRule{
			Host: hostname,
			TrafficPolicy: &istiov1alpha3spec.TrafficPolicy{
				Tls: &istiov1alpha3spec.ClientTLSSettings{
					// TODO(ilackarms): currently we set all DRs to mTLS
					// in the future we'll want to make this configurable
					// https://github.com/solo-io/service-mesh-hub/issues/790
					Mode: istiov1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
				},
			},
		},
	}
}

func (t *translator) trafficPolicyToResourceIds(
	trafficPolicy []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
) []ezkube.ResourceId {
	var resourceIds []ezkube.ResourceId
	for _, policy := range trafficPolicy {
		resourceIds = append(resourceIds, policy.Ref)
	}
	return resourceIds
}
