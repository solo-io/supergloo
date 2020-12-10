package traffictarget

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/authorizationpolicy"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/destinationrule"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/virtualservice"
	istioUtils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./istio_traffic_target_translator.go -destination mocks/istio_traffic_target_translator.go

// the VirtualService translator translates a TrafficTarget into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given TrafficTarget.
	// returns nil if no VirtualService or DestinationRule is required for the TrafficTarget (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		outputs istio.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx                   context.Context
	destinationRules      destinationrule.Translator
	virtualServices       virtualservice.Translator
	authorizationPolicies authorizationpolicy.Translator
}

func NewTranslator(
	ctx context.Context,
	clusterDomains hostutils.ClusterDomainRegistry,
	decoratorFactory decorators.Factory,
	trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	failoverServices v1alpha2sets.FailoverServiceSet,
) Translator {
	return &translator{
		ctx:                   ctx,
		destinationRules:      destinationrule.NewTranslator(clusterDomains, decoratorFactory, trafficTargets, failoverServices),
		virtualServices:       virtualservice.NewTranslator(clusterDomains, decoratorFactory),
		authorizationPolicies: authorizationpolicy.NewTranslator(),
	}
}

// translate the appropriate resources for the given TrafficTarget.
func (t *translator) Translate(
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	outputs istio.Builder,
	reporter reporting.Reporter,
) {
	// only translate istio trafficTargets
	if !t.isIstioTrafficTarget(t.ctx, trafficTarget, in.Meshes()) {
		return
	}

	if istioUtils.IsIstioInternal(trafficTarget) {
		contextutils.LoggerFrom(t.ctx).Debugf("skipping istio internal services %v", trafficTarget.Name)
		return
	}

	// Pass nil sourceMeshInstallation to translate VirtualService local to trafficTarget
	vs := t.virtualServices.Translate(in, trafficTarget, nil, reporter)
	dr := t.destinationRules.Translate(t.ctx, in, trafficTarget, nil, reporter)
	ap := t.authorizationPolicies.Translate(in, trafficTarget, reporter)

	outputs.AddVirtualServices(vs)
	outputs.AddDestinationRules(dr)
	outputs.AddAuthorizationPolicies(ap)
}

func (t *translator) isIstioTrafficTarget(
	ctx context.Context,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	allMeshes discoveryv1alpha2sets.MeshSet,
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
	return mesh.Spec.GetIstio() != nil
}
