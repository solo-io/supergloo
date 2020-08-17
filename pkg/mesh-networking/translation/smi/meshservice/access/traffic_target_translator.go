package access

import (
	"context"
	"fmt"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/rotisserie/eris"
	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/pkg/ezkube"
)

const

//go:generate mockgen -source ./traffic_target_translator.go -destination mocks/traffic_target_translator.go

// the VirtualService translator translates a MeshService into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService for the given MeshService.
	// returns nil if no VirtualService is required for the MeshService (i.e. if no VirtualService features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot MeshServiceSet contains the given MeshService.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		reporter reporting.Reporter,
	) ([]*smiaccessv1alpha2.TrafficTarget, []*smispecsv1alpha3.HTTPRouteGroup)
}

func NewTranslator() Translator {
	return &translator{}
}

type translator struct {
}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	meshService *discoveryv1alpha2.MeshService,
	reporter reporting.Reporter,
) ([]*smiaccessv1alpha2.TrafficTarget, []*smispecsv1alpha3.HTTPRouteGroup) {

	kubeService := meshService.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil, nil
	}

	var trafficTargets []*smiaccessv1alpha2.TrafficTarget
	var httpRouteGroups []*smispecsv1alpha3.HTTPRouteGroup

	backingWorkloads := workloadutils.FindBackingMeshWorkloads(meshService, in.MeshWorkloads())
	for _, ap := range meshService.Status.GetAppliedAccessPolicies() {

		if backingWorkloads.Length() == 0 {
			reporter.ReportAccessPolicyToMeshService(
				meshService,
				ap.GetRef(),
				eris.New("Could not determine ServiceAccount target for MeshService as no backing workloads exist"),
			)
			continue
		} else if backingWorkloads.Length() > 1 {
			reporter.ReportAccessPolicyToMeshService(
				meshService,
				ap.GetRef(),
				eris.New("Could not determine ServiceAccount target for MeshService as more than 1 workloads exist"),
			)
			continue
		}

		var trafficTargetsByAp []*smiaccessv1alpha2.TrafficTarget

		backingWorkload := backingWorkloads.List()[0]
		trafficTarget := &smiaccessv1alpha2.TrafficTarget{
			ObjectMeta: metautils.TranslatedObjectMeta(
				meshService.Spec.GetKubeService().Ref,
				meshService.Annotations,
			),
			Spec: smiaccessv1alpha2.TrafficTargetSpec{
				Destination: smiaccessv1alpha2.IdentityBindingSubject{
					Kind:      "ServiceAccount",
					Namespace: backingWorkload.Spec.GetKubernetes().GetController().GetNamespace(),
					Name:      backingWorkload.Spec.GetKubernetes().GetServiceAccountName(),
				},
			},
		}

		for _, sourceSelector := range ap.GetSpec().GetSourceSelector() {
			if sourceSelector.GetKubeIdentityMatcher() != nil {
				reporter.ReportAccessPolicyToMeshService(
					meshService,
					ap.GetRef(),
					eris.New("SMI does not support KubeIdentityMatcher for IdentitySelectors"),
				)
			}

			for _, ref := range sourceSelector.GetKubeServiceAccountRefs().GetServiceAccounts() {
				if ref.GetClusterName() != kubeService.GetRef().GetClusterName() {
					reporter.ReportAccessPolicyToMeshService(
						meshService,
						ap.GetRef(),
						eris.New("SMI identity sources cannot be multi cluster"),
					)
					continue
				}
				trafficTarget.Spec.Sources = append(trafficTarget.Spec.Sources,
					smiaccessv1alpha2.IdentityBindingSubject{
						Kind:      "ServiceAccount",
						Name:      ref.GetName(),
						Namespace: ref.GetNamespace(),
					},
				)
			}
		}

		// Append the ap ref to the name as each ap gets it's own traffic target
		trafficTarget.Name += fmt.Sprintf("-%s", kubeValidName(ap.GetRef()))

		if len(ap.GetSpec().GetAllowedPorts()) > 1 {
			// Add a traffic target per port
			for _, port := range ap.GetSpec().GetAllowedPorts() {
				ttByPort := trafficTarget.DeepCopy()
				ttByPort.Name += fmt.Sprintf("-%d", port)
				intPort := int(port)
				ttByPort.Spec.Destination.Port = &intPort
				trafficTargetsByAp = append(trafficTargetsByAp, ttByPort)
			}

		} else {
			// Add the default traffic target
			trafficTargetsByAp = append(trafficTargetsByAp, trafficTarget)
		}

		httpMatch := smispecsv1alpha3.HTTPMatch{
			Name:      kubeValidName(ap.GetRef()),
			Methods:   methodsToString(ap.GetSpec().GetAllowedMethods()),
			// Need to default to * or OSM does not route at all
			PathRegex: constants.RegexMatchAll,
		}

		var httpMatches []smispecsv1alpha3.HTTPMatch

		if len(ap.GetSpec().GetAllowedPaths()) > 1 {
			for idx, path := range ap.GetSpec().GetAllowedPaths() {
				matchByPath := httpMatch.DeepCopy()
				matchByPath.Name += fmt.Sprintf("-%d", idx)
				matchByPath.PathRegex = path
				httpMatches = append(httpMatches, httpMatch)
			}
		} else {
			httpMatches = append(httpMatches, httpMatch)
		}

		routeGroup := &smispecsv1alpha3.HTTPRouteGroup{
			ObjectMeta: metautils.TranslatedObjectMeta(
				meshService.Spec.GetKubeService().Ref,
				meshService.Annotations,
			),
			Spec: smispecsv1alpha3.HTTPRouteGroupSpec{
				Matches: httpMatches,
			},
		}
		// Append the ap ref to the name as each ap gets it's own route group
		routeGroup.Name += fmt.Sprintf("-%s", kubeValidName(ap.GetRef()))

		// add all of the http matches to all of the traffic targets
		for _, tt := range trafficTargetsByAp {
			rule := smiaccessv1alpha2.TrafficTargetRule{
				Kind: "HTTPRouteGroup",
				Name: routeGroup.GetName(),
			}
			for _, match := range routeGroup.Spec.Matches {
				rule.Matches = append(rule.Matches, match.Name)
			}
			tt.Spec.Rules = append(tt.Spec.Rules, rule)
		}

		httpRouteGroups = append(httpRouteGroups, routeGroup)
		trafficTargets = append(trafficTargets, trafficTargetsByAp...)
	}
	return trafficTargets, httpRouteGroups
}

func kubeValidName(id ezkube.ResourceId) string {
	return id.GetName() + "." + id.GetNamespace()
}

func methodsToString(methods []types.HttpMethodValue) []string {
	// If no method(s) has been specified, need to default to all or OSM doesn't work
	if len(methods) == 0 {
		return []string{string(smispecsv1alpha3.HTTPRouteMethodAll)}
	}
	var result []string
	for _, method := range methods {
		result = append(result, method.String())
	}
	return result
}
