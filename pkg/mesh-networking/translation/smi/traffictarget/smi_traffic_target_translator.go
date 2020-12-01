package traffictarget

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/smi"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/traffictarget/access"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/traffictarget/split"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
)

//go:generate mockgen -source ./smi_traffic_target_translator.go -destination mocks/smi_traffic_target_translator.go

// the VirtualService translator translates a TrafficTarget into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given TrafficTarget.
	// returns nil if no VirtualService or DestinationRule is required for the TrafficTarget (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the smi
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		outputs smi.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	trafficSplit  split.Translator
	trafficTarget access.Translator
}

func NewTranslator(tsTranslator split.Translator, ttTranslator access.Translator) Translator {
	return &translator{
		trafficSplit:  tsTranslator,
		trafficTarget: ttTranslator,
	}
}

// translate the appropriate resources for the given TrafficTarget.
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	outputs smi.Builder,
	reporter reporting.Reporter,
) {
	var appliedTrafficPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy
	for _, tp := range trafficTarget.Status.GetAppliedTrafficPolicies() {
		if len(tp.GetSpec().GetTrafficShift().GetDestinations()) > 0 {
			appliedTrafficPolicy = tp
			break
		}
	}

	// Translate TrafficSplit for TrafficTarget, can be nil if non-kube service or no applied traffic policy
	if trafficSplit := t.trafficSplit.Translate(ctx, in, trafficTarget, reporter); trafficSplit != nil {
		if appliedTrafficPolicy != nil {
			// Append the applied traffic policy as the parent to the traffic split
			metautils.AppendParent(ctx, trafficSplit, appliedTrafficPolicy.GetRef(), v1alpha2.TrafficPolicy{}.GVK())
		}

		outputs.AddTrafficSplits(trafficSplit)
	}

	// Translate output TrafficTargets and HttpRouteGroups for discovered TrafficTarget,
	// can be nil if non-kube service
	trafficTargets, httpRouteGroups := t.trafficTarget.Translate(ctx, in, trafficTarget, reporter)
	if trafficTarget != nil {
		if appliedTrafficPolicy != nil {
			// Append the applied traffic policy as the parent to each traffic target
			for _, tt := range trafficTargets {
				metautils.AppendParent(ctx, tt, appliedTrafficPolicy.GetRef(), v1alpha2.TrafficPolicy{}.GVK())
			}
		}

		outputs.AddTrafficTargets(trafficTargets...)
	}
	if httpRouteGroups != nil {
		// Append the applied traffic policy as the parent to each http route group
		if appliedTrafficPolicy != nil {
			for _, rg := range httpRouteGroups {
				metautils.AppendParent(ctx, rg, appliedTrafficPolicy.GetRef(), v1alpha2.TrafficPolicy{}.GVK())
			}
		}

		outputs.AddHTTPRouteGroups(httpRouteGroups...)
	}
}
