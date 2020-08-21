package traffictarget

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/smi"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/access"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/split"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./smi_traffic_target_translator.go -destination mocks/smi_traffic_target_translator.go

// the VirtualService translator translates a TrafficTarget into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given TrafficTarget.
	// returns nil if no VirtualService or DestinationRule is required for the TrafficTarget (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the smi
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		meshService *discoveryv1alpha2.TrafficTarget,
		outputs smi.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	trafficSplit  split.Translator
	trafficTarget access.Translator
}

func NewTranslator(tsTranslator split.Translator, ttTranslator access.Translator) Translator {
	return &translator{
		trafficSplit:  tsTranslator,
		trafficTarget: ttTranslator,
	}
}

// translate the appropriate resources for the given TrafficTarget.
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	meshService *discoveryv1alpha2.TrafficTarget,
	outputs smi.Builder,
	reporter reporting.Reporter,
) {
	// only translate istio meshServices
	if !t.isSmiTrafficTarget(ctx, meshService, in.Meshes()) {
		return
	}

	ts := t.trafficSplit.Translate(ctx, in, meshService, reporter)
	outputs.AddTrafficSplits(ts)

	tt, hgr := t.trafficTarget.Translate(ctx, in, meshService, reporter)
	outputs.AddTrafficTargets(tt...)
	outputs.AddHTTPRouteGroups(hgr...)
}

func (t *translator) isSmiTrafficTarget(
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
	return mesh.Spec.GetSmiEnabled()
}
