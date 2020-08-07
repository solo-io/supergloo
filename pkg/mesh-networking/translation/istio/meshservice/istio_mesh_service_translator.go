package meshservice

import (
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/authorizationpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/destinationrule"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/virtualservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
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
	destinationRules      destinationrule.Translator
	virtualServices       virtualservice.Translator
	authorizationPolicies authorizationpolicy.Translator
}

func NewTranslator(clusterDomains hostutils.ClusterDomainRegistry, decoratorFactory decorators.Factory) Translator {
	return &translator{
		destinationRules:      destinationrule.NewTranslator(clusterDomains, decoratorFactory),
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

	vs := t.virtualServices.Translate(in, meshService, reporter)
	dr := t.destinationRules.Translate(in, meshService, reporter)
	ap := t.authorizationPolicies.Translate(in, meshService, reporter)

	outputs.AddVirtualServices(vs)
	outputs.AddDestinationRules(dr)
	outputs.AddAuthorizationPolicies(ap)
}
