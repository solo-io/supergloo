package traffictarget

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/traffictarget/authorizationpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/traffictarget/destinationrule"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/traffictarget/virtualservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
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
	allMeshes             v1alpha2sets.MeshSet
	destinationRules      destinationrule.Translator
	virtualServices       virtualservice.Translator
	authorizationPolicies authorizationpolicy.Translator
}

func NewTranslator(
	ctx context.Context,
	allMeshes v1alpha2sets.MeshSet,
	clusterDomains hostutils.ClusterDomainRegistry,
	decoratorFactory decorators.Factory,
	trafficTargets v1alpha2sets.TrafficTargetSet,
) Translator {
	return &translator{
		ctx:                   ctx,
		allMeshes:             allMeshes,
		destinationRules:      destinationrule.NewTranslator(clusterDomains, decoratorFactory, trafficTargets),
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
	// only translate istio meshServices
	if !t.isIstioMeshService(t.ctx, trafficTarget, t.allMeshes) {
		return
	}

	vs := t.virtualServices.Translate(in, trafficTarget, reporter)
	dr := t.destinationRules.Translate(in, trafficTarget, reporter)
	ap := t.authorizationPolicies.Translate(in, trafficTarget, reporter)

	outputs.AddVirtualServices(vs)
	outputs.AddDestinationRules(dr)
	outputs.AddAuthorizationPolicies(ap)
}

func (t *translator) isIstioMeshService(
	ctx context.Context,
	meshService *discoveryv1alpha2.TrafficTarget,
	allMeshes v1alpha2sets.MeshSet,
) bool {
	meshRef := meshService.Spec.Mesh
	if meshRef == nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: meshService %v missing mesh ref", sets.Key(meshService))
		return false
	}
	mesh, err := allMeshes.Find(meshRef)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: could not find mesh %v for meshService %v", sets.Key(meshRef), sets.Key(meshService))
		return false
	}
	return mesh.Spec.GetIstio() != nil
}
