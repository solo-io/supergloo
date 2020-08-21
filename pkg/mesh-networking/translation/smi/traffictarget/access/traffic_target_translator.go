package access

import (
	"context"
	"fmt"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/rotisserie/eris"
	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	sksets "github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
)

//go:generate mockgen -source ./traffic_target_translator.go -destination mocks/traffic_target_translator.go

// the SMI Access translator translates a TrafficTarget into sets of SMI access resources.
type Translator interface {
	// Translate translates the appropriate TrafficTargets and HTTPRoutesGroups for the given TrafficTarget.
	// returns empty lists if none are required
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot TrafficTargetSet contains the given TrafficTarget.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		meshService *discoveryv1alpha2.TrafficTarget,
		reporter reporting.Reporter,
	) ([]*smiaccessv1alpha2.TrafficTarget, []*smispecsv1alpha3.HTTPRouteGroup)
}

var (
	NoServiceAccountError = eris.New("Could not determine ServiceAccount target for TrafficTarget as no backing" +
		" workloads exist")

	CouldNotDetermineServiceAccountError = func(total int) error {
		return eris.Errorf("Could not determine ServiceAccount target for TrafficTarget as workloads belong to "+
			"%d service accounts", total)
	}
)

func NewTranslator() Translator {
	return &translator{}
}

type translator struct{}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	target *discoveryv1alpha2.TrafficTarget,
	reporter reporting.Reporter,
) ([]*smiaccessv1alpha2.TrafficTarget, []*smispecsv1alpha3.HTTPRouteGroup) {
	logger := contextutils.LoggerFrom(ctx).With(zap.String("translator", "access"))
	kubeService := target.Spec.GetKubeService()

	if kubeService == nil {
		logger.Debugf("non kubernetes mesh service %s found, skipping", sksets.TypedKey(target))
		return nil, nil
	}

	var trafficTargets []*smiaccessv1alpha2.TrafficTarget
	var httpRouteGroups []*smispecsv1alpha3.HTTPRouteGroup

	backingWorkloads := workloadutils.FindBackingWorkloads(target.Spec.GetKubeService(), in.Workloads())
	for _, ap := range target.Status.GetAppliedAccessPolicies() {

		if len(backingWorkloads) == 0 {
			reporter.ReportAccessPolicyToTrafficTarget(
				target,
				ap.GetRef(),
				NoServiceAccountError,
			)
			continue
		}

		// Check workloads for all possible service accounts, if they belong to more than 1, report an error
		serviceAccounts := sets.NewString()
		for _, workload := range backingWorkloads {
			kubeWorkload := workload.Spec.GetKubernetes()
			if kubeWorkload == nil {
				continue
			}
			serviceAccounts.Insert(kubeWorkload.GetServiceAccountName())
		}

		if serviceAccounts.Len() > 1 {
			reporter.ReportAccessPolicyToTrafficTarget(
				target,
				ap.GetRef(),
				CouldNotDetermineServiceAccountError(serviceAccounts.Len()),
			)
			continue
		}

		var trafficTargetsByAp []*smiaccessv1alpha2.TrafficTarget

		backingWorkload := backingWorkloads[0]
		trafficTarget := &smiaccessv1alpha2.TrafficTarget{
			ObjectMeta: metautils.TranslatedObjectMeta(
				target.Spec.GetKubeService().Ref,
				target.Annotations,
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
				reporter.ReportAccessPolicyToTrafficTarget(
					target,
					ap.GetRef(),
					eris.New("SMI does not support KubeIdentityMatcher for IdentitySelectors"),
				)
			}

			for _, ref := range sourceSelector.GetKubeServiceAccountRefs().GetServiceAccounts() {
				if ref.GetClusterName() != kubeService.GetRef().GetClusterName() {
					reporter.ReportAccessPolicyToTrafficTarget(
						target,
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
		trafficTarget.Name += fmt.Sprintf(".%s", t.kubeValidName(ap.GetRef()))

		if len(ap.GetSpec().GetAllowedPorts()) > 1 {
			// Add a traffic target per port
			for _, port := range ap.GetSpec().GetAllowedPorts() {
				ttByPort := trafficTarget.DeepCopy()
				ttByPort.Name += fmt.Sprintf(".%d", port)
				intPort := int(port)
				ttByPort.Spec.Destination.Port = &intPort
				trafficTargetsByAp = append(trafficTargetsByAp, ttByPort)
			}

		} else {
			// Add the default traffic target
			trafficTargetsByAp = append(trafficTargetsByAp, trafficTarget)
		}

		httpMatch := smispecsv1alpha3.HTTPMatch{
			Name:    t.kubeValidName(ap.GetRef()),
			Methods: t.methodsToString(ap.GetSpec().GetAllowedMethods()),
			// Need to default to * or OSM does not route at all
			PathRegex: constants.RegexMatchAll,
		}

		var httpMatches []smispecsv1alpha3.HTTPMatch

		if len(ap.GetSpec().GetAllowedPaths()) > 1 {
			for idx, path := range ap.GetSpec().GetAllowedPaths() {
				matchByPath := httpMatch.DeepCopy()
				matchByPath.Name += fmt.Sprintf(".%d", idx)
				matchByPath.PathRegex = path
				httpMatches = append(httpMatches, httpMatch)
			}
		} else {
			httpMatches = append(httpMatches, httpMatch)
		}

		routeGroup := &smispecsv1alpha3.HTTPRouteGroup{
			ObjectMeta: metautils.TranslatedObjectMeta(
				target.Spec.GetKubeService().Ref,
				target.Annotations,
			),
			Spec: smispecsv1alpha3.HTTPRouteGroupSpec{
				Matches: httpMatches,
			},
		}
		// Append the ap ref to the name as each ap gets it's own route group
		routeGroup.Name += fmt.Sprintf(".%s", t.kubeValidName(ap.GetRef()))

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

func (t *translator) kubeValidName(id ezkube.ResourceId) string {
	return id.GetName() + "." + id.GetNamespace()
}

func (t *translator) methodsToString(methods []types.HttpMethodValue) []string {
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
