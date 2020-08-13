package split

import (
	"context"
	"fmt"

	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smislpitv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha3"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

//go:generate mockgen -source ./traffic_split_translator.go -destination mocks/traffic_split_translator.go

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
	) (*smislpitv1alpha3.TrafficSplit, *smispecsv1alpha3.HTTPRouteGroup)
}

type registerFieldFunc func(backends *[]smislpitv1alpha3.TrafficSplitBackend) error

func NewUnsupportedFeatureError(resource ezkube.ResourceId, fieldName, reason string) error {
	return &UnsupportedFeatureError{
		resource:  resource,
		fieldName: fieldName,
		reason:    reason,
	}
}

type UnsupportedFeatureError struct {
	resource  ezkube.ResourceId
	fieldName string
	reason    string
}

func (u *UnsupportedFeatureError) Error() string {
	return fmt.Sprintf(
		"Unsupported feature %s used on resource %T <%s>. %s",
		u.fieldName,
		u.resource,
		sets.Key(u.resource),
		u.reason,
	)
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
) (*smislpitv1alpha3.TrafficSplit, *smispecsv1alpha3.HTTPRouteGroup) {
	kubeService := meshService.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	trafficSplit := &smislpitv1alpha3.TrafficSplit{
		ObjectMeta: metautils.TranslatedObjectMeta(
			meshService.Spec.GetKubeService().Ref,
			meshService.Annotations,
		),
		Spec: smislpitv1alpha3.TrafficSplitSpec{
			// TODO: Fix this key
			Service:  fmt.Sprintf("%s.%s", kubeService.GetRef().GetName(), kubeService.GetRef().GetNamespace()),
			Backends: nil,
			Matches:  nil,
		},
	}

	httpRoute := &smispecsv1alpha3.HTTPRouteGroup{}

	trafficTarget := &smiaccessv1alpha2.TrafficTarget{}
	// register the owners of the traffic split fields
	trafficSplitFields := fieldutils.NewOwnershipRegistry()
	fieldFunc := t.registerFieldFunc()

	for _, tp := range meshService.Status.GetAppliedTrafficPolicies() {
		if tp.Spec.GetCorsPolicy() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"CorsPolicy",
				"Smi does not support cors policy",
			))
		}
		if tp.Spec.GetFaultInjection() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"FaultInjection",
				"Smi does not support fault injection",
			))
		}
		if tp.Spec.GetHeaderManipulation() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"HeaderManipulation",
				"Smi does not support header manipulation",
			))
		}
		if tp.Spec.GetMirror() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"Mirror",
				"Smi does not support request mirroring",
			))
		}
		if tp.Spec.GetRetries() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"Mirror",
				"Smi does not support retries",
			))
		}
		if tp.Spec.GetRequestTimeout() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"RequestTimeout",
				"Smi does not support request timeout",
			))
		}
		tp.Spec.GetTrafficShift()
	}

}

func (t *translator) buildRouteMatches(
	tp *v1.ObjectRef,
	matchers []*v1alpha2.TrafficPolicySpec_HttpMatcher,
) ([]smispecsv1alpha3.HTTPMatch, error) {
	var result []smispecsv1alpha3.HTTPMatch
	for idx, matcher := range matchers {
		if len(matcher.GetQueryParameters()) != 0 {
			return nil, NewUnsupportedFeatureError(
				tp,
				"HttpMatcher.QueryParameters",
				"Smi does not support query parameter matching",
			)
		}

		if matcher.GetPrefix() != "" {
			return nil, NewUnsupportedFeatureError(
				tp,
				"HttpMatcher.Prefix",
				"Smi does not support prefix path matching",
			)
		}
		if matcher.GetExact() != "" {
			return nil, NewUnsupportedFeatureError(
				tp,
				"HttpMatcher.Exact",
				"Smi does not support exact path matching",
			)
		}

		httpMatch := smispecsv1alpha3.HTTPMatch{
			// create a unique key for this HTTP route match name
			Name:      fmt.Sprintf("%s_%d", sets.Key(tp), idx),
			PathRegex: matcher.GetRegex(),
			// Initialize just in case
			Headers: map[string]string{},
		}

		if matcher.GetMethod() != nil {
			// If matcher is present, translate it to string and set it
			httpMatch.Methods = []string{matcher.GetMethod().GetMethod().String()}
		} else {
			// Otherwise use *
			httpMatch.Methods = []string{string(smispecsv1alpha3.HTTPRouteMethodAll)}
		}

		for _, header := range matcher.GetHeaders() {

			if header.GetInvertMatch() {
				return nil, NewUnsupportedFeatureError(
					tp,
					"HttpMatcher.Headers.Invert",
					"Smi does not support inverted header matching",
				)
			}

			if header.GetRegex() {
				return nil, NewUnsupportedFeatureError(
					tp,
					"HttpMatcher.Headers.Regex",
					"Smi does not support regex header matching",
				)
			}

			// Add all of the headers to the header match map
			httpMatch.Headers[header.GetName()] = header.GetValue()
		}
		result = append(result, httpMatch)
	}
	return result, nil
}

func (t *translator) buildBackends() []smislpitv1alpha3.TrafficSplitBackend {

}

// construct the callback for registering fields in the virtual service
func (t *translator) registerFieldFunc(
	trafficSplitFields fieldutils.FieldOwnershipRegistry,
	trafficSplit *smislpitv1alpha3.TrafficSplit,
	policyRef ezkube.ResourceId,
) registerFieldFunc {
	return func(backends *[]smislpitv1alpha3.TrafficSplitBackend) error {
		if err := trafficSplitFields.RegisterFieldOwnership(
			trafficSplit,
			backends,
			[]ezkube.ResourceId{policyRef},
			&v1alpha2.TrafficPolicy{},
			0,
		); err != nil {
			return err
		}
		return nil
	}
}
