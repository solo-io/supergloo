package meshservice

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/authorizationpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/destinationrule"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/virtualservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./istio_mesh_service_translator.go -destination mocks/istio_mesh_service_translator.go

// the VirtualService translator translates a MeshService into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given MeshService.
	// returns nil if no VirtualService or DestinationRule is required for the MeshService (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		outputs output.Builder,
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

func NewTranslator(ctx context.Context, allMeshes v1alpha2sets.MeshSet, clusterDomains hostutils.ClusterDomainRegistry, decoratorFactory decorators.Factory, meshServices v1alpha2sets.MeshServiceSet) Translator {
	return &translator{
		ctx:                   ctx,
		allMeshes:             allMeshes,
		destinationRules:      destinationrule.NewTranslator(clusterDomains, decoratorFactory, meshServices),
		virtualServices:       virtualservice.NewTranslator(clusterDomains, decoratorFactory),
		authorizationPolicies: authorizationpolicy.NewTranslator(),
	}
}

// translate the appropriate resources for the given MeshService.
func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha2.MeshService,
	outputs output.Builder,
	reporter reporting.Reporter,
) {
	// only translate istio meshServices
	if !t.isIstioMeshService(t.ctx, meshService, t.allMeshes) {
		return
	}

	vs := t.virtualServices.Translate(in, meshService, reporter)
	dr := t.destinationRules.Translate(in, meshService, reporter)
	ap := t.authorizationPolicies.Translate(in, meshService, reporter)

	outputs.AddVirtualServices(vs)
	outputs.AddDestinationRules(dr)
	outputs.AddAuthorizationPolicies(ap)
}

func (t *translator) isIstioMeshService(
	ctx context.Context,
	meshService *discoveryv1alpha2.MeshService,
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
