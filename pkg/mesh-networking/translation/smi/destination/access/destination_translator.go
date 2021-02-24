package access

import (
	"context"
	"fmt"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/rotisserie/eris"
	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	sksets "github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
)

//go:generate mockgen -source ./destination_translator.go -destination mocks/destination_translator.go

// the SMI Access translator translates a Destination into sets of SMI access resources.
type Translator interface {
	// Translate translates the appropriate Destinations and HTTPRoutesGroups for the given Destination.
	// returns empty lists if none are required
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot DestinationSet contains the given Destination.
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
		destination *discoveryv1.Destination,
		reporter reporting.Reporter,
	) ([]*smiaccessv1alpha2.TrafficTarget, []*smispecsv1alpha3.HTTPRouteGroup)
}

var (
	NoServiceAccountError = eris.New("Could not determine ServiceAccount target for Destination as no backing" +
		" workloads exist")

	CouldNotDetermineServiceAccountError = func(total int) error {
		return eris.Errorf("Could not determine ServiceAccount target for Destination as workloads belong to "+
			"%d service accounts", total)
	}
)

func NewTranslator() Translator {
	return &translator{}
}

type translator struct{}

func (t *translator) Translate(
	ctx context.Context,
	in input.LocalSnapshot,
	destination *discoveryv1.Destination,
	reporter reporting.Reporter,
) ([]*smiaccessv1alpha2.TrafficTarget, []*smispecsv1alpha3.HTTPRouteGroup) {
	logger := contextutils.LoggerFrom(ctx).With(zap.String("translator", "access"))
	kubeService := destination.Spec.GetKubeService()

	if kubeService == nil {
		logger.Debugf("non kubernetes Destination %s found, skipping", sksets.TypedKey(destination))
		return nil, nil
	}

	var trafficTargets []*smiaccessv1alpha2.TrafficTarget
	var httpRouteGroups []*smispecsv1alpha3.HTTPRouteGroup

	backingWorkloads := workloadutils.FindBackingWorkloads(destination.Spec.GetKubeService(), in.Workloads())
	for _, ap := range destination.Status.GetAppliedAccessPolicies() {

		if len(backingWorkloads) == 0 {
			reporter.ReportAccessPolicyToDestination(
				destination,
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
			reporter.ReportAccessPolicyToDestination(
				destination,
				ap.GetRef(),
				CouldNotDetermineServiceAccountError(serviceAccounts.Len()),
			)
			continue
		}

		var trafficTargetsByAp []*smiaccessv1alpha2.TrafficTarget

		backingWorkload := backingWorkloads[0]
		trafficTarget := &smiaccessv1alpha2.TrafficTarget{
			ObjectMeta: metautils.TranslatedObjectMeta(
				destination.Spec.GetKubeService().Ref,
				destination.Annotations,
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
				reporter.ReportAccessPolicyToDestination(
					destination,
					ap.GetRef(),
					eris.New("SMI does not support KubeIdentityMatcher for IdentitySelectors"),
				)
			}

			for _, ref := range sourceSelector.GetKubeServiceAccountRefs().GetServiceAccounts() {
				if ref.GetClusterName() != kubeService.GetRef().GetClusterName() {
					reporter.ReportAccessPolicyToDestination(
						destination,
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

		// Append the ap ref to the name as each ap gets it's own Destination
		trafficTarget.Name += fmt.Sprintf(".%s", t.kubeValidName(ap.GetRef()))

		if len(ap.GetSpec().GetAllowedPorts()) > 1 {
			// Add a Destination per port
			for _, port := range ap.GetSpec().GetAllowedPorts() {
				ttByPort := trafficTarget.DeepCopy()
				ttByPort.Name += fmt.Sprintf(".%d", port)
				intPort := int(port)
				ttByPort.Spec.Destination.Port = &intPort
				trafficTargetsByAp = append(trafficTargetsByAp, ttByPort)
			}

		} else {
			// Add the default Destination
			trafficTargetsByAp = append(trafficTargetsByAp, trafficTarget)
		}

		methods := ap.GetSpec().GetAllowedMethods()
		// If no method(s) has been specified, need to default to all or OSM doesn't work
		if len(methods) == 0 {
			methods = []string{string(smispecsv1alpha3.HTTPRouteMethodAll)}
		}

		httpMatch := smispecsv1alpha3.HTTPMatch{
			Name:    t.kubeValidName(ap.GetRef()),
			Methods: methods,
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
				destination.Spec.GetKubeService().Ref,
				destination.Annotations,
			),
			Spec: smispecsv1alpha3.HTTPRouteGroupSpec{
				Matches: httpMatches,
			},
		}
		// Append the ap ref to the name as each ap gets it's own route group
		routeGroup.Name += fmt.Sprintf(".%s", t.kubeValidName(ap.GetRef()))

		// add all of the http matches to all of the Destinations
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

		// Append the Destination as a parent to each output route group
		metautils.AppendParent(ctx, routeGroup, ap.GetRef(), v1.AccessPolicy{}.GVK())
		metautils.AppendParent(ctx, trafficTarget, ap.GetRef(), v1.AccessPolicy{}.GVK())

		httpRouteGroups = append(httpRouteGroups, routeGroup)
		trafficTargets = append(trafficTargets, trafficTargetsByAp...)
	}
	return trafficTargets, httpRouteGroups
}

func (t *translator) kubeValidName(id ezkube.ResourceId) string {
	return id.GetName() + "." + id.GetNamespace()
}
