package traffictarget

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/settingsutils"

	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
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
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
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
		in input.LocalSnapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		outputs istio.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx                   context.Context
	userSupplied          input.RemoteSnapshot
	destinationRules      destinationrule.Translator
	virtualServices       virtualservice.Translator
	authorizationPolicies authorizationpolicy.Translator
}

func NewTranslator(
	ctx context.Context,
	userSupplied input.RemoteSnapshot,
	clusterDomains hostutils.ClusterDomainRegistry,
	decoratorFactory decorators.Factory,
	trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	failoverServices v1alpha2sets.FailoverServiceSet,
) Translator {
	var existingVirtualServices v1alpha3sets.VirtualServiceSet
	var existingDestinationRules v1alpha3sets.DestinationRuleSet
	if userSupplied != nil {
		existingVirtualServices = userSupplied.VirtualServices()
		existingDestinationRules = userSupplied.DestinationRules()
	}

	return &translator{
		ctx:                   ctx,
		destinationRules:      destinationrule.NewTranslator(settingsutils.SettingsFromContext(ctx), existingDestinationRules, clusterDomains, decoratorFactory, trafficTargets, failoverServices),
		virtualServices:       virtualservice.NewTranslator(existingVirtualServices, clusterDomains, decoratorFactory),
		authorizationPolicies: authorizationpolicy.NewTranslator(),
	}
}

// translate the appropriate resources for the given TrafficTarget.
func (t *translator) Translate(
	in input.LocalSnapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	outputs istio.Builder,
	reporter reporting.Reporter,
) {
	// only translate istio trafficTargets
	if !t.isIstioTrafficTarget(t.ctx, trafficTarget, in.Meshes()) {
		return
	}

	// Translate VirtualServices for TrafficTargets, can be nil if there is no service or applied traffic policies
	// Pass nil sourceMeshInstallation to translate VirtualService local to trafficTarget
	vs := t.virtualServices.Translate(t.ctx, in, trafficTarget, nil, reporter)
	// Append the traffic target as a parent to the virtual service
	metautils.AppendParent(t.ctx, vs, trafficTarget, trafficTarget.GVK())
	outputs.AddVirtualServices(vs)
	// Translate DestinationRules for TrafficTargets, can be nil if there is no service or applied traffic policies
	dr := t.destinationRules.Translate(t.ctx, in, trafficTarget, nil, reporter)
	// Append the traffic target as a parent to the destination rule
	metautils.AppendParent(t.ctx, dr, trafficTarget, trafficTarget.GVK())
	outputs.AddDestinationRules(dr)
	// Translate AuthorizationPolicies for TrafficTargets, can be nil if there is no service or applied traffic policies
	ap := t.authorizationPolicies.Translate(in, trafficTarget, reporter)
	// Append the traffic target as a parent to the authorization policy
	metautils.AppendParent(t.ctx, ap, trafficTarget, trafficTarget.GVK())
	outputs.AddAuthorizationPolicies(ap)
}

func (t *translator) isIstioTrafficTarget(
	ctx context.Context,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	allMeshes discoveryv1alpha2sets.MeshSet,
) bool {
	meshRef := trafficTarget.Spec.GetMesh()
	if meshRef == nil {
		contextutils.LoggerFrom(ctx).Debugf("trafficTarget %v has no mesh ref - is not istio trafficTarget", sets.Key(trafficTarget))
		return false
	}
	mesh, err := allMeshes.Find(meshRef)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: could not find mesh %v for trafficTarget %v", sets.Key(meshRef), sets.Key(trafficTarget))
		return false
	}
	return mesh.Spec.GetIstio() != nil
}
