package destinationrule

import (
	"reflect"

	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/mesh-networking/plugins"
	"github.com/solo-io/smh/pkg/mesh-networking/reporting"
	destinationruleplugin "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/destinationrule/plugins"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
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
		meshService *discoveryv1alpha1.MeshService,
		reporter reporting.Reporter,
	) *istiov1alpha3.DestinationRule
}

type translator struct {
	clusterDomains hostutils.ClusterDomainRegistry
	pluginFactory  plugins.Factory
}

func NewTranslator(clusterDomains hostutils.ClusterDomainRegistry, pluginFactory plugins.Factory) Translator {
	return &translator{clusterDomains: clusterDomains, pluginFactory: pluginFactory}
}

// translate the appropriate DestinationRUle for the given MeshService.
// returns nil if no DestinationRule is required for the MeshService (i.e. if no DestinationRule features are required, such as subsets).
// The input snapshot MeshServiceSet contains n the
func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha1.MeshService,
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
	drPlugins := t.pluginFactory.MakePlugins(plugins.Parameters{
		ClusterDomains: t.clusterDomains,
		Snapshot:       in,
	})

	for _, policy := range meshService.Status.AppliedTrafficPolicies {
		registerField := registerFieldFunc(destinationRuleFields, destinationRule, policy.Ref)
		for _, plugin := range drPlugins {

			if trafficPolicyPlugin, ok := plugin.(destinationruleplugin.TrafficPolicyPlugin); ok {
				if err := trafficPolicyPlugin.ProcessTrafficPolicy(
					policy,
					meshService,
					&destinationRule.Spec,
					registerField,
				); err != nil {
					reporter.ReportTrafficPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", plugin.PluginName()))
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
	policyRef ezkube.ResourceId,
) plugins.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.Equals(fieldVal, val) {
			return nil
		}
		if err := destinationRuleFields.RegisterFieldOwner(
			destinationRule,
			fieldPtr,
			policyRef,
			&v1alpha1.TrafficPolicy{},
			0, //TODO(ilackarms): priority
		); err != nil {
			return err
		}
		return nil
	}
}

func (t *translator) initializeDestinationRule(meshService *discoveryv1alpha1.MeshService) *istiov1alpha3.DestinationRule {
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
