package traffictarget

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/smi"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	smitraffictarget "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./osm_traffic_target_translator.go -destination mocks/osm_traffic_target_translator.go

// the VirtualService translator translates a TrafficTarget into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given TrafficTarget.
	// returns nil if no VirtualService or DestinationRule is required for the TrafficTarget (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the smi
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
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

// translate the appropriate resources for the given TrafficTarget.
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	outputs smi.Builder,
	reporter reporting.Reporter,
) {
	// only translate osm trafficTargets
	if !t.isOSMTrafficTarget(ctx, trafficTarget, in.Meshes()) {
		return
	}

	t.smiTranslator.Translate(ctx, in, trafficTarget, outputs, reporter)
}

func (t *translator) isOSMTrafficTarget(
	ctx context.Context,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	allMeshes v1alpha2sets.MeshSet,
) bool {
	meshRef := trafficTarget.Spec.Mesh
	if meshRef == nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: trafficTarget %v missing mesh ref", sets.Key(trafficTarget))
		return false
	}
	mesh, err := allMeshes.Find(meshRef)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: could not find mesh %v for trafficTarget %v", sets.Key(meshRef), sets.Key(trafficTarget))
		return false
	}

	return mesh.Spec.GetOsm() != nil
}
