package mesh

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
)

//go:generate mockgen -source ./appmesh_mesh_translator.go -destination mocks/appmesh_mesh_translator.go

// Translator translates Meshes into appmesh outputs.
type Translator interface {
	// Translate translates the appropriate resources to apply the VirtualMesh to the given Mesh.
	// Output resources will be added to the appmesh.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		mesh *discoveryv1alpha2.Mesh,
		outputs appmesh.Builder,
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
	outputs appmesh.Builder,
	reporter reporting.Reporter,
) {

	appmeshMesh := mesh.Spec.GetAwsAppMesh()
	if appmeshMesh == nil {
		return
	}

	for _, cluster := range appmeshMesh.Clusters {
		outputs.AddCluster(cluster)
	}
}
