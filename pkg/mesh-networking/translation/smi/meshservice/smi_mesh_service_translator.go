package meshservice

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice/access"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice/split"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./smi_mesh_service_translator.go -destination mocks/smi_mesh_service_translator.go

// the VirtualService translator translates a MeshService into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given MeshService.
	// returns nil if no VirtualService or DestinationRule is required for the MeshService (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		meshService *discoveryv1alpha2.MeshService,
		outputs output.Builder,
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

// translate the appropriate resources for the given MeshService.
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	meshService *discoveryv1alpha2.MeshService,
	outputs output.Builder,
	reporter reporting.Reporter,
) {
	// only translate istio meshServices
	if !t.isSmiMeshService(ctx, meshService, in.Meshes()) {
		return
	}

	ts := t.trafficSplit.Translate(ctx, in, meshService, reporter)
	outputs.AddTrafficSplits(ts)

	tt, hgr := t.trafficTarget.Translate(ctx, in, meshService, reporter)
	outputs.AddTrafficTargets(tt...)
	outputs.AddHTTPRouteGroups(hgr...)
}

func (t *translator) isSmiMeshService(
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
	return mesh.Spec.GetSmiEnabled()
}
