package meshservice

import (
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/authorizationpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/destinationrule"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/virtualservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiov1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// outputs of translating a single MeshService
type Outputs struct {
	VirtualService      *networkingv1alpha3.VirtualService
	DestinationRule     *networkingv1alpha3.DestinationRule
	AuthorizationPolicy *istiov1beta1.AuthorizationPolicy
}

// the VirtualService translator translates a MeshService into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given MeshService.
	// returns nil if no VirtualService or DestinationRule is required for the MeshService (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		reporter reporting.Reporter,
	) Outputs
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
	reporter reporting.Reporter,
) Outputs {

	vs := t.virtualServices.Translate(in, meshService, reporter)
	dr := t.destinationRules.Translate(in, meshService, reporter)
	ap := t.authorizationPolicies.Translate(in, meshService, reporter)

	return Outputs{
		VirtualService:      vs,
		DestinationRule:     dr,
		AuthorizationPolicy: ap,
	}
}
