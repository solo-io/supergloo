package traffictarget

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/errors"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./appmesh_traffic_target_translator.go -destination mocks/appmesh_traffic_target_translator.go

// Translator translator translates a TrafficTarget into a equivalent appmesh config.
type Translator interface {
	// Output resources will configure the underlying appmesh mesh.
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		outputs appmesh.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct{}

func NewTranslator() Translator {
	return &translator{}
}

// translate the appropriate resources for the given TrafficTarget.
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	outputs appmesh.Builder,
	reporter reporting.Reporter,
) {
	// only translate appmesh trafficTargets
	if !isAppmeshTrafficTarget(ctx, trafficTarget, in.Meshes()) {
		return
	}

	kubeService := trafficTarget.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	var appliedTrafficPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy

	for _, tp := range trafficTarget.Status.GetAppliedTrafficPolicies() {
		validate(tp, trafficTarget, reporter)

		// If there is no traffic shifting, skip the rest of the translation
		if len(tp.GetSpec().GetTrafficShift().GetDestinations()) == 0 {
			continue
		} else if appliedTrafficPolicy != nil {
			// Each smi traffic target can only have a single applied traffic policy
			reporter.ReportTrafficPolicyToTrafficTarget(
				trafficTarget,
				tp.GetRef(),
				eris.New("SMI only supports one TrafficSplit per service, multiple found"),
			)
			continue
		}

		appliedTrafficPolicy = tp
	}

}

func isAppmeshTrafficTarget(
	ctx context.Context,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	allMeshes v1alpha2sets.MeshSet,
) bool {
	meshRef := trafficTarget.Spec.Mesh
	if meshRef == nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: trafficTarget %v missing mesh ref", sets.Key(trafficTarget))
		return false
	}
	mesh, err := allMeshes.Find(meshRef)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: could not find mesh %v for trafficTarget %v", sets.Key(meshRef), sets.Key(trafficTarget))
		return false
	}

	return mesh.Spec.GetAwsAppMesh() != nil
}

func validate(
	tp *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	reporter reporting.Reporter,
) {
	getMessage := func(feature string) string {
		return fmt.Sprintf("Service Mesh Hub does not support %s for AppMesh", feature)
	}

	if tp.GetSpec().GetCorsPolicy() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"CorsPolicy",
			getMessage("CorsPolicy"),
		))
	}
	if tp.GetSpec().GetFaultInjection() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"FaultInjection",
			getMessage("FaultInjection"),
		))
	}
	if tp.GetSpec().GetHeaderManipulation() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"HeaderManipulation",
			getMessage("HeaderManipulation"),
		))
	}
	if tp.GetSpec().GetMirror() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"Mirror",
			getMessage("Mirror"),
		))
	}
	if tp.GetSpec().GetRetries() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"Retries",
			getMessage("Retries"),
		))
	}
	if tp.GetSpec().GetRequestTimeout() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"RequestTimeout",
			getMessage("RequestTimeout"),
		))
	}
	if tp.GetSpec().GetSourceSelector() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"SourceSelector",
			getMessage("SourceSelector"),
		))
	}
}
