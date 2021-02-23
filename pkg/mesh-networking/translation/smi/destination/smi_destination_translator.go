package destination

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/smi"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination/access"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination/split"
)

//go:generate mockgen -source ./smi_destination_translator.go -destination mocks/smi_destination_translator.go

// translates a Destination into a OSM resources.
type Translator interface {
	// Translate translates OSM resources for the given Destination.
	// Output resources will be added to the smi output snapshot
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
		destination *discoveryv1alpha2.Destination,
		outputs smi.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	trafficSplit split.Translator
	destination  access.Translator
}

func NewTranslator(tsTranslator split.Translator, ttTranslator access.Translator) Translator {
	return &translator{
		trafficSplit: tsTranslator,
		destination:  ttTranslator,
	}
}

// translate the appropriate resources for the given Destination.
func (t *translator) Translate(
	ctx context.Context,
	in input.LocalSnapshot,
	destination *discoveryv1alpha2.Destination,
	outputs smi.Builder,
	reporter reporting.Reporter,
) {
	// Translate TrafficSplit for Destination, can be nil if non-kube service or no applied TrafficPolicy
	trafficSplit := t.trafficSplit.Translate(ctx, in, destination, reporter)

	// Translate output Destinations and HttpRouteGroups for discovered Destination
	trafficTargets, httpRouteGroups := t.destination.Translate(ctx, in, destination, reporter)

	outputs.AddTrafficSplits(trafficSplit)
	outputs.AddTrafficTargets(trafficTargets...)
	outputs.AddHTTPRouteGroups(httpRouteGroups...)
}
