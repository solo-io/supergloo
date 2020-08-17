package split

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smislpitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
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
	) *smislpitv1alpha2.TrafficSplit
}

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
) *smislpitv1alpha2.TrafficSplit {
	kubeService := meshService.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	trafficSplit := &smislpitv1alpha2.TrafficSplit{
		ObjectMeta: metautils.TranslatedObjectMeta(
			meshService.Spec.GetKubeService().Ref,
			meshService.Annotations,
		),
		Spec: smislpitv1alpha2.TrafficSplitSpec{
			// TODO: Fix this key
			Service:  fmt.Sprintf("%s.%s", kubeService.GetRef().GetName(), kubeService.GetRef().GetNamespace()),
			Backends: nil,
		},
	}

	for _, tp := range meshService.Status.GetAppliedTrafficPolicies() {
		if tp.GetSpec().GetCorsPolicy() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"CorsPolicy",
				"Smi does not support cors policy",
			))
		}
		if tp.GetSpec().GetFaultInjection() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"FaultInjection",
				"Smi does not support fault injection",
			))
		}
		if tp.GetSpec().GetHeaderManipulation() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"HeaderManipulation",
				"Smi does not support header manipulation",
			))
		}
		if tp.GetSpec().GetMirror() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"Mirror",
				"Smi does not support request mirroring",
			))
		}
		if tp.GetSpec().GetRetries() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"Mirror",
				"Smi does not support retries",
			))
		}
		if tp.GetSpec().GetRequestTimeout() != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), NewUnsupportedFeatureError(
				tp.GetRef(),
				"RequestTimeout",
				"Smi does not support request timeout",
			))
		}

		// If there is no traffic shifting, skip the rest of the translation
		if len(tp.GetSpec().GetTrafficShift().GetDestinations()) == 0 {
			continue
		} else if len(trafficSplit.Spec.Backends) != 0 {
			// Each smi mesh service can only have a single applied traffic policy
			// TODO(EItanya): clearer error
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), eris.New("too many owners"))
		}

		backends, err := t.buildBackends(tp.GetRef(), tp.Spec.GetTrafficShift(), kubeService)
		if err != nil {
			reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), err)
		}

		trafficSplit.Spec.Backends = backends

		// Only create the route group if there are any matchers
		if len(tp.GetSpec().GetHttpRequestMatchers()) == 0 {
			continue
		}

		// httpMatchers, err := t.buildRouteMatches(tp.GetRef(), tp.Spec.GetHttpRequestMatchers())
		// if err != nil {
		// 	reporter.ReportTrafficPolicyToMeshService(meshService, tp.GetRef(), err)
		// }

		//		// httpRouteGroup = &smispecsv1alpha3.HTTPRouteGroup{
		// 	ObjectMeta: metautils.TranslatedObjectMeta(
		// 		meshService.Spec.GetKubeService().Ref,
		// 		meshService.Annotations,
		// 	),
		// 	// set the route group matchers to the current
		// 	Spec: smispecsv1alpha3.HTTPRouteGroupSpec{
		// 		Matches: httpMatchers,
		// 	},
		// }
	}
	return trafficSplit
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
				fmt.Sprintf("HttpMatcher[%d].Prefix", idx),
				"Smi does not support prefix path matching",
			)
		}
		if matcher.GetExact() != "" {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("HttpMatcher[%d].Exact", idx),
				"Smi does not support exact path matching",
			)
		}

		httpMatch := smispecsv1alpha3.HTTPMatch{
			// create a unique key for this HTTP route match name
			Name:      fmt.Sprintf("%s_%d", sets.TypedKey(tp), idx),
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

		for headerIdx, header := range matcher.GetHeaders() {

			if header.GetInvertMatch() {
				return nil, NewUnsupportedFeatureError(
					tp,
					fmt.Sprintf("HttpMatcher[%d].Headers[%d].Invert", idx, headerIdx),
					"Smi does not support inverted header matching",
				)
			}

			if header.GetRegex() {
				return nil, NewUnsupportedFeatureError(
					tp,
					fmt.Sprintf("HttpMatcher[%d].Headers[%d].Regex", idx, headerIdx),
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

func (t *translator) buildBackends(
	tp *v1.ObjectRef,
	multiDest *v1alpha2.TrafficPolicySpec_MultiDestination,
	meshKubeService *discoveryv1alpha2.MeshServiceSpec_KubeService,
) ([]smislpitv1alpha2.TrafficSplitBackend, error) {
	var result []smislpitv1alpha2.TrafficSplitBackend
	for idx, dest := range multiDest.GetDestinations() {
		backend := smislpitv1alpha2.TrafficSplitBackend{
			Weight: int(dest.GetWeight()),
		}
		kubeService := dest.GetKubeService()
		if kubeService == nil {
			return nil, eris.New("no desination type found")
		}

		if len(kubeService.GetSubset()) != 0 {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("TrafficShift.Destination[%d].Subest", idx),
				"Smi does not support subset routing",
			)
		}

		if kubeService.GetPort() != 0 {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("TrafficShift.Destination[%d].Port", idx),
				"Smi does not support specifying a service port for traffic shifting",
			)
		}

		if kubeService.GetClusterName() != meshKubeService.GetRef().GetClusterName() {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("TrafficShift.Destination[%d].Cluster", idx),
				"Smi does not currently support multi cluster traffic shifting",
			)
		}

		backend.Service = fmt.Sprintf("%s.%s", kubeService.GetName(), kubeService.GetNamespace())
		result = append(result, backend)
	}
	return result, nil
}
