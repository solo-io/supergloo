package destination

import (
	"context"

	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/authorizationpolicy"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/federation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/virtualservice"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/settingsutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate mockgen -source ./istio_destination_translator.go -destination mocks/istio_destination_translator.go

// the VirtualService translator translates a Destination into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given Destination.
	// returns nil if no VirtualService or DestinationRule is required for the Destination (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		eventObjs map[schema.GroupVersionKind][]ezkube.ResourceId,
		in input.LocalSnapshot,
		destination *discoveryv1.Destination,
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
	federation            federation.Translator
}

func NewTranslator(
	ctx context.Context,
	userSupplied input.RemoteSnapshot,
	clusterDomains hostutils.ClusterDomainRegistry,
	decoratorFactory decorators.Factory,
) Translator {
	var existingVirtualServices v1alpha3sets.VirtualServiceSet
	var existingDestinationRules v1alpha3sets.DestinationRuleSet
	if userSupplied != nil {
		existingVirtualServices = userSupplied.VirtualServices()
		existingDestinationRules = userSupplied.DestinationRules()
	}

	virtualServiceTranslator := virtualservice.NewTranslator(existingVirtualServices, clusterDomains, decoratorFactory)
	destinationRuleTranslator := destinationrule.NewTranslator(settingsutils.SettingsFromContext(ctx), existingDestinationRules, clusterDomains, decoratorFactory)

	return &translator{
		ctx:                   ctx,
		virtualServices:       virtualServiceTranslator,
		destinationRules:      destinationRuleTranslator,
		authorizationPolicies: authorizationpolicy.NewTranslator(),
		federation:            federation.NewTranslator(virtualServiceTranslator, destinationRuleTranslator),
	}
}

// translate the appropriate resources for the given Destination.
func (t *translator) Translate(
	eventObjs map[schema.GroupVersionKind][]ezkube.ResourceId,
	in input.LocalSnapshot,
	destination *discoveryv1.Destination,
	outputs istio.Builder,
	reporter reporting.Reporter,
) {
	// only translate istio Destinations
	if !t.isIstioDestination(t.ctx, destination, in.Meshes()) {
		return
	}

	if t.virtualServices.ShouldTranslate(destination, eventObjs) {
		// Translate VirtualServices for Destinations, can be nil if there is no service or applied traffic policies
		// Pass nil sourceMeshInstallation to translate VirtualService local to destination
		vs := t.virtualServices.Translate(t.ctx, in, destination, nil, reporter)
		outputs.AddVirtualServices(vs)
	}

	if t.destinationRules.ShouldTranslate(destination, eventObjs) {
		// Translate DestinationRules for Destinations, can be nil if there is no service or applied traffic policies
		dr := t.destinationRules.Translate(t.ctx, in, destination, nil, reporter)
		outputs.AddDestinationRules(dr)
	}

	if t.authorizationPolicies.ShouldTranslate(destination, eventObjs) {
		// Translate AuthorizationPolicies for Destinations, can be nil if there is no service or applied traffic policies
		ap := t.authorizationPolicies.Translate(t.ctx, in, destination, reporter)
		outputs.AddAuthorizationPolicies(ap)
	}

	if t.federation.ShouldTranslate(destination, eventObjs) {
		serviceEntries, virtualServices, destinationRules := t.federation.Translate(t.ctx, in, destination, reporter)
		outputs.AddServiceEntries(serviceEntries...)
		outputs.AddVirtualServices(virtualServices...)
		outputs.AddDestinationRules(destinationRules...)
	}
}

func (t *translator) isIstioDestination(
	ctx context.Context,
	destination *discoveryv1.Destination,
	allMeshes discoveryv1sets.MeshSet,
) bool {
	meshRef := destination.Spec.Mesh
	if meshRef == nil {
		if _, ok := destination.Spec.Type.(*discoveryv1.DestinationSpec_KubeService_); ok {
			// Is KubeService, MeshRef is required
			contextutils.LoggerFrom(ctx).Errorf("internal error: KubeService destination %v missing mesh ref", sets.Key(destination))
		} else {
			// Not KubeService, MeshRef not required
			// TODO: Publish this as a metric in the future
		}
		return false
	}
	mesh, err := allMeshes.Find(meshRef)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: could not find mesh %v for destination %v", sets.Key(meshRef), sets.Key(destination))
		return false
	}
	return mesh.Spec.GetIstio() != nil
}
