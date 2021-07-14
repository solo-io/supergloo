package split

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	smislpitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

//go:generate mockgen -source ./traffic_split_translator.go -destination mocks/traffic_split_translator.go

// the TrafficSplit Translator translates a Destination into a TrafficSplit.
type Translator interface {
	// Translate translates the appropriate TrafficSplit for the given Destination.
	// returns nil if no TrafficSplit is required
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot DestinationSet contains the given Destination.
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
		destination *discoveryv1.Destination,
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

type translator struct{}

func (t *translator) Translate(
	ctx context.Context,
	_ input.LocalSnapshot,
	destination *discoveryv1.Destination,
	reporter reporting.Reporter,
) *smislpitv1alpha2.TrafficSplit {
	kubeService := destination.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	var appliedTrafficPolicy *v1.AppliedTrafficPolicy

	for _, tp := range destination.Status.GetAppliedTrafficPolicies() {
		validate(tp, destination, reporter)

		// If there is no traffic shifting, skip the rest of the translation
		if len(tp.GetSpec().GetPolicy().GetTrafficShift().GetDestinations()) == 0 {
			continue
		} else if appliedTrafficPolicy != nil {
			// Each smi traffic target can only have a single applied TrafficPolicy
			reporter.ReportTrafficPolicyToDestination(
				destination,
				tp.GetRef(),
				eris.New("SMI only supports one TrafficSplit per service, multiple found"),
			)
			continue
		}

		appliedTrafficPolicy = tp
	}

	// If no TrafficPolicy has been applied, return nil
	if appliedTrafficPolicy == nil {
		return nil
	}

	trafficSplit := &smislpitv1alpha2.TrafficSplit{
		ObjectMeta: metautils.TranslatedObjectMeta(
			destination.Spec.GetKubeService().Ref,
			destination.Annotations,
		),
		Spec: smislpitv1alpha2.TrafficSplitSpec{
			// SMI expects this to be <namespace>.<name> of the target service
			Service: fmt.Sprintf("%s.%s", kubeService.GetRef().GetName(), kubeService.GetRef().GetNamespace()),
			// Will populate this late
			Backends: nil,
		},
	}

	backends, err := buildBackends(appliedTrafficPolicy.GetRef(), appliedTrafficPolicy.Spec.GetPolicy().GetTrafficShift(), kubeService)
	if err != nil {
		reporter.ReportTrafficPolicyToDestination(destination, appliedTrafficPolicy.GetRef(), err)
	}

	trafficSplit.Spec.Backends = backends

	// Append the applied TrafficPolicy as the parent to the traffic split
	// Done here to avoid duplicating the logic to find the applied TrafficPolicy in the SMI Destination translator
	metautils.AppendParent(ctx, trafficSplit, appliedTrafficPolicy.GetRef(), v1.TrafficPolicy{}.GVK())

	return trafficSplit
}

func validate(
	tp *v1.AppliedTrafficPolicy,
	destination *discoveryv1.Destination,
	reporter reporting.Reporter,
) {
	if tp.GetSpec().GetPolicy().GetCorsPolicy() != nil {
		reporter.ReportTrafficPolicyToDestination(destination, tp.GetRef(), NewUnsupportedFeatureError(
			tp.GetRef(),
			"CorsPolicy",
			"SMI does not support cors policy",
		))
	}
	if tp.GetSpec().GetPolicy().GetFaultInjection() != nil {
		reporter.ReportTrafficPolicyToDestination(destination, tp.GetRef(), NewUnsupportedFeatureError(
			tp.GetRef(),
			"FaultInjection",
			"SMI does not support fault injection",
		))
	}
	if tp.GetSpec().GetPolicy().GetHeaderManipulation() != nil {
		reporter.ReportTrafficPolicyToDestination(destination, tp.GetRef(), NewUnsupportedFeatureError(
			tp.GetRef(),
			"HeaderManipulation",
			"SMI does not support header manipulation",
		))
	}
	if tp.GetSpec().GetPolicy().GetMirror() != nil {
		reporter.ReportTrafficPolicyToDestination(destination, tp.GetRef(), NewUnsupportedFeatureError(
			tp.GetRef(),
			"Mirror",
			"SMI does not support request mirroring",
		))
	}
	if tp.GetSpec().GetPolicy().GetRetries() != nil {
		reporter.ReportTrafficPolicyToDestination(destination, tp.GetRef(), NewUnsupportedFeatureError(
			tp.GetRef(),
			"Retries",
			"SMI does not support retries",
		))
	}
	if tp.GetSpec().GetPolicy().GetRequestTimeout() != nil {
		reporter.ReportTrafficPolicyToDestination(destination, tp.GetRef(), NewUnsupportedFeatureError(
			tp.GetRef(),
			"RequestTimeout",
			"SMI does not support request timeout",
		))
	}
	if tp.GetSpec().GetSourceSelector() != nil {
		reporter.ReportTrafficPolicyToDestination(destination, tp.GetRef(), NewUnsupportedFeatureError(
			tp.GetRef(),
			"SourceSelector",
			"SMI does not support source selectors for traffic policies",
		))
	}
}

func buildBackends(
	tp *skv2corev1.ObjectRef,
	multiDest *v1.TrafficPolicySpec_Policy_MultiDestination,
	meshKubeService *discoveryv1.DestinationSpec_KubeService,
) ([]smislpitv1alpha2.TrafficSplitBackend, error) {
	var result []smislpitv1alpha2.TrafficSplitBackend
	for idx, dest := range multiDest.GetDestinations() {
		backend := smislpitv1alpha2.TrafficSplitBackend{
			Weight: int(dest.GetWeight()),
		}
		kubeService := dest.GetKubeService()
		if kubeService == nil {
			return nil, eris.Errorf("SMI traffic split only supports Kube destinations, found %T", dest.GetDestinationType())
		}

		if len(kubeService.GetSubset()) != 0 {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("TrafficShift.Destination[%d].Subest", idx),
				"SMI does not support subset routing",
			)
		}

		if kubeService.GetPort() != 0 {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("TrafficShift.Destination[%d].Port", idx),
				"SMI does not support specifying a service port for traffic shifting",
			)
		}

		if kubeService.GetClusterName() != meshKubeService.GetRef().GetClusterName() {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("TrafficShift.Destination[%d].Cluster", idx),
				"SMI does not currently support multi cluster traffic shifting",
			)
		}

		if kubeService.GetNamespace() != meshKubeService.GetRef().GetNamespace() {
			return nil, NewUnsupportedFeatureError(
				tp,
				fmt.Sprintf("TrafficShift.Destination[%d].Namespace", idx),
				"SMI does not support traffic split destinations in different namespaces",
			)
		}

		backend.Service = kubeService.GetName()
		result = append(result, backend)
	}
	return result, nil
}
