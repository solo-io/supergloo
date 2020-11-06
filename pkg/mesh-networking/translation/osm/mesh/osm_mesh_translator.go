package mesh

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/smi"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
)

//go:generate mockgen -source ./osm_mesh_translator.go -destination mocks/osm_mesh_translator.go

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate resources to apply the VirtualMesh to the given Mesh.
	// Output resources will be added to the smi.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		mesh *discoveryv1alpha2.Mesh,
		outputs smi.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx context.Context
}

func NewTranslator() Translator {
	return &translator{}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	mesh *discoveryv1alpha2.Mesh,
	outputs smi.Builder,
	reporter reporting.Reporter,
) {

	osmMesh := mesh.Spec.GetOsm()
	if osmMesh == nil {
		return
	}

	outputs.AddCluster(osmMesh.GetInstallation().GetCluster())
}
