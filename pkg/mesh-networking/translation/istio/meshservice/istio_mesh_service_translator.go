package meshservice

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/smh/pkg/mesh-networking/reporting"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/destinationrule"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// outputs of translating a single MeshService
type Outputs struct {
	VirtualService  *istiov1alpha3.VirtualService
	DestinationRule *istiov1alpha3.DestinationRule
}

// the VirtualService translator translates a MeshService into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given MeshService.
	// returns nil if no VirtualService or DestinationRule is required for the MeshService (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		meshService *discoveryv1alpha1.MeshService,
		reporter reporting.Reporter,
	) Outputs
}

type translator struct {
	destinationRules destinationrule.Translator
	virtualServices  virtualservice.Translator
}

func NewTranslator(clusterDomains hostutils.ClusterDomainRegistry, decoratorFactory decorators.Factory) Translator {
	return &translator{
		destinationRules: destinationrule.NewTranslator(clusterDomains, decoratorFactory),
		virtualServices:  virtualservice.NewTranslator(clusterDomains, decoratorFactory),
	}
}

// translate the appropriate resources for the given MeshService.
func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha1.MeshService,
	reporter reporting.Reporter,
) Outputs {

	vs := t.virtualServices.Translate(in, meshService, reporter)
	dr := t.destinationRules.Translate(in, meshService, reporter)

	return Outputs{
		VirtualService:  vs,
		DestinationRule: dr,
	}
}
