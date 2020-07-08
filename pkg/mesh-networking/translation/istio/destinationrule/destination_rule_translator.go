package destinationrule

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/destinationrule/plugin"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"reflect"
	"sort"
	"strings"
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
		reporter reporter.Reporter,
	) *istiov1alpha3.DestinationRule
}

type translator struct {
	clusterDomains hostutils.ClusterDomainRegistry
	pluginFactory  plugin.Factory
}

func NewTranslator(clusterDomains hostutils.ClusterDomainRegistry, pluginFactory plugin.Factory) Translator {
	return &translator{clusterDomains: clusterDomains, pluginFactory: pluginFactory}
}

// translate the appropriate DestinationRUle for the given MeshService.
// returns nil if no DestinationRule is required for the MeshService (i.e. if no DestinationRule features are required, such as subsets).
// The input snapshot MeshServiceSet contains n the
func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha1.MeshService,
	reporter reporter.Reporter,
) *istiov1alpha3.DestinationRule {
	plugins := t.pluginFactory.MakePlugins(in)
	destinationRule := t.initializeDestinationRule(meshService)

	for _, plug := range plugins {
		if simplePlugin, ok := plug.(plugin.SimplePlugin); ok {
			simplePlugin.Process(meshService, destinationRule)
		}
	}
	for _, policy := range meshService.Status.AppliedTrafficPolicies {
		for _, plug := range plugins {
			if trafficPolicyPlugin, ok := plug.(plugin.TrafficPolicyPlugin); ok {
				if err := trafficPolicyPlugin.ProcessTrafficPolicy(policy.Spec, meshService, destinationRule); err != nil {
					reporter.ReportTrafficPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", plug.PluginName()))
				}
			}
		}
	}

	for _, policy := range meshService.Status.AppliedAccessPolicies {
		for _, plug := range plugins {
			if accessPolicyPlugin, ok := plug.(plugin.AccessPolicyPlugin); ok {
				if err := accessPolicyPlugin.ProcessAccessPolicy(policy.Spec, meshService, destinationRule); err != nil {
					reporter.ReportAccessPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", plug.PluginName()))
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

func (t *translator) initializeDestinationRule(meshService *discoveryv1alpha1.MeshService) *istiov1alpha3.DestinationRule {
	meta := metautils.TranslatedObjectMeta(
		meshService.Spec.KubeService.Ref,
		meshService.Annotations,
	)
	hostname := t.clusterDomains.GetServiceLocalFQDN(meshService.Spec.KubeService.Ref)
	subsets := buildRequiredSubsets(meshService)

	return &istiov1alpha3.DestinationRule{
		ObjectMeta: meta,
		Spec: istiov1alpha3spec.DestinationRule{
			Host:    hostname,
			Subsets: subsets,
		},
	}
}

func buildRequiredSubsets(meshService *discoveryv1alpha1.MeshService) []*istiov1alpha3spec.Subset {
	var uniqueSubsets []map[string]string
	appendUniqueSubset := func(subsetLabels map[string]string) {
		for _, subset := range uniqueSubsets {
			if reflect.DeepEqual(subset, subsetLabels) {
				return
			}
		}
		uniqueSubsets = append(uniqueSubsets, subsetLabels)
	}

	for _, policy := range meshService.Status.AppliedTrafficPolicies {
		for _, destination := range policy.GetSpec().GetTrafficShift().GetDestinations() {
			if subsetLabels := destination.GetSubset(); len(subsetLabels) > 0 {
				appendUniqueSubset(subsetLabels)
			}
		}
	}

	var subsets []*istiov1alpha3spec.Subset
	for _, subsetLabels := range uniqueSubsets {
		subsets = append(subsets, &istiov1alpha3spec.Subset{
			Name:   SubsetName(subsetLabels),
			Labels: subsetLabels,
		})
	}

	return subsets
}

// used in VirtualService translator as well
func SubsetName(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	var keys []string
	for key, val := range labels {
		keys = append(keys, key+"-"+val)
	}
	sort.Strings(keys)
	return kubeutils.SanitizeNameV2(strings.Join(keys, "_"))
}
