package authorizationpolicy

import (
	"reflect"

	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/pkg/ezkube"
	securityv1beta1spec "istio.io/api/security/v1beta1"
	typesv1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// the AuthorizationPolicy translator translates a MeshService into a AuthorizationPolicy.
type Translator interface {
	// Translate translates the appropriate AuthorizationPolicy for the given MeshService.
	// returns nil if no AuthorizationPolicy is required for the MeshService (i.e. if no AuthorizationPolicy features are required, such access control).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot MeshServiceSet contains the given MeshService.
	Translate(
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		reporter reporting.Reporter,
	) *securityv1beta1.AuthorizationPolicy
}

type translator struct {
	decoratorFactory decorators.Factory
}

func NewTranslator(decoratorFactory decorators.Factory) *translator {
	return &translator{decoratorFactory: decoratorFactory}
}

func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha2.MeshService,
	reporter reporting.Reporter,
) *securityv1beta1.AuthorizationPolicy {
	kubeService := meshService.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(harveyxia): non kube services currently unsupported
		return nil
	}

	authPolicy := t.initializeAuthorizationPolicy(meshService)
	// register the owners of the AuthorizationPolicy fields
	authPolicyFields := fieldutils.NewOwnershipRegistry()
	apDecorators := t.decoratorFactory.MakeDecorators(decorators.Parameters{
		Snapshot: in,
	})

	// Apply decorators which map a single applicable TrafficPolicy to a field on the AuthorizationPolicy.
	for _, policy := range meshService.Status.AppliedAccessPolicies {
		registerField := registerFieldFunc(authPolicyFields, authPolicy, []ezkube.ResourceId{policy.Ref})
		for _, decorator := range apDecorators {

			if authPolicyDecorator, ok := decorator.(accesspolicy.AuthorizationPolicyDecorator); ok {
				if err := authPolicyDecorator.ApplyToAuthorizationPolicy(
					policy,
					meshService,
					&authPolicy.Spec,
					registerField,
				); err != nil {
					reporter.ReportTrafficPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", decorator.DecoratorName()))
				}
			}
		}
	}

	if len(authPolicy.Spec.Rules) == 0 {
		// no need to create this AuthorizationPolicy as it has no effect
		return nil
	}
	return authPolicy
}

// construct the callback for registering fields in the virtual service
func registerFieldFunc(
	authPolicyFields fieldutils.FieldOwnershipRegistry,
	authPolicy *securityv1beta1.AuthorizationPolicy,
	policyRefs []ezkube.ResourceId,
) decorators.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.Equals(fieldVal, val) {
			return nil
		}
		if err := authPolicyFields.RegisterFieldOwnership(
			authPolicy,
			fieldPtr,
			policyRefs,
			&v1alpha2.AccessPolicy{},
			0, //TODO(harveyxia): priority
		); err != nil {
			return err
		}
		return nil
	}
}

func (t *translator) initializeAuthorizationPolicy(meshService *discoveryv1alpha2.MeshService) *securityv1beta1.AuthorizationPolicy {
	meta := metautils.TranslatedObjectMeta(
		meshService.Spec.GetKubeService().Ref,
		meshService.Annotations,
	)
	authPolicy := &securityv1beta1.AuthorizationPolicy{
		ObjectMeta: meta,
		Spec: securityv1beta1spec.AuthorizationPolicy{
			Selector: &typesv1beta1.WorkloadSelector{
				MatchLabels: meshService.Spec.GetKubeService().WorkloadSelectorLabels,
			},
			Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
		},
	}
	return authPolicy
}
