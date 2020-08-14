package access


import (
	"context"

	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
)

//go:generate mockgen -source ./traffic_target_translator.go -destination mocks/traffic_target_translator.go

// the VirtualService translator translates a MeshService into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService for the given MeshService.
	// returns nil if no VirtualService is required for the MeshService (i.e. if no VirtualService features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot MeshServiceSet contains the given MeshService.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		reporter reporting.Reporter,
	) (*smiaccessv1alpha2.TrafficTarget, *v1alpha3.HTTPRouteGroup)
}

func NewTranslator() Translator {
	return &translator{}
}

type translator struct {
}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	meshService *discoveryv1alpha2.MeshService,
	reporter reporting.Reporter,
) (*smiaccessv1alpha2.TrafficTarget, *v1alpha3.HTTPRouteGroup) {
	panic("not implemented")
	// for _, tp := range meshService.Status.GetAppliedTrafficPolicies() {
	// 	tp.Spec.GetTrafficShift()
	// }
}
