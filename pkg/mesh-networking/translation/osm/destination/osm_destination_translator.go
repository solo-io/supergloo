package destination

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/smi"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	smitraffictarget "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./osm_destination_translator.go -destination mocks/osm_destination_translator.go

// the VirtualService translator translates a Destination into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given Destination.
	// returns nil if no VirtualService or DestinationRule is required for the Destination (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the smi
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
	smiTranslator smitraffictarget.Translator
}

func NewTranslator(smiTranslator smitraffictarget.Translator) Translator {
	return &translator{
		smiTranslator: smiTranslator,
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
	// only translate osm Destinations
	if !t.isOSMDestination(ctx, destination, in.Meshes()) {
		return
	}

	t.smiTranslator.Translate(ctx, in, destination, outputs, reporter)
}

func (t *translator) isOSMDestination(
	ctx context.Context,
	destination *discoveryv1alpha2.Destination,
	allMeshes v1alpha2sets.MeshSet,
) bool {
	meshRef := destination.Spec.Mesh
	if meshRef == nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: destination %v missing mesh ref", sets.Key(destination))
		return false
	}
	mesh, err := allMeshes.Find(meshRef)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: could not find mesh %v for destination %v", sets.Key(meshRef), sets.Key(destination))
		return false
	}

	return mesh.Spec.GetOsm() != nil
}
