package traffictarget

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/appmesh/traffictarget/virtualrouter"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/appmesh/traffictarget/virtualservice"
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
	// Only translate appmesh TrafficTargets.
	if !isAppmeshTrafficTarget(ctx, trafficTarget, in.Meshes()) {
		return
	}

	// AppMesh doesn't support all policies; report those which aren't implemented.
	for _, tp := range trafficTarget.Status.GetAppliedTrafficPolicies() {
		report(tp, trafficTarget, reporter)
	}

	virtualRouter := virtualrouter.NewVirtualRouterTranslator().Translate(ctx, in, trafficTarget, reporter)
	virtualServices := virtualservice.NewVirtualServiceTranslator().Translate(ctx, in, trafficTarget, virtualRouter, reporter)

	outputs.AddVirtualServices(virtualServices...)
	outputs.AddVirtualRouters(virtualRouter)
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

func report(
	tp *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	reporter reporting.Reporter,
) {
	getMessage := func(feature string) string {
		return fmt.Sprintf("Service Mesh Hub does not support %s for AppMesh", feature)
	}

	// TODO joekelley add mTLS here

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
