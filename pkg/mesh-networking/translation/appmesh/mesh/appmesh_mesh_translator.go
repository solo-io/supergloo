package mesh

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/appmesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
)

//go:generate mockgen -source ./appmesh_mesh_translator.go -destination mocks/appmesh_mesh_translator.go

// Translator translates DiscoveryMeshGlooSoloIov1Alpha2Meshes into appmesh outputs.
type Translator interface {
	// Translate translates the appropriate resources to apply the VirtualMesh to the given Mesh.
	// Output resources will be added to the appmesh.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
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
	in input.LocalSnapshot,
	mesh *discoveryv1alpha2.Mesh,
	outputs appmesh.Builder,
	reporter reporting.Reporter,
) {

	appmeshMesh := mesh.Spec.GetAwsAppMesh()
	if appmeshMesh == nil {
		return
	}

	for cluster := range appmeshMesh.ClusterMeshResources {
		outputs.AddCluster(cluster)
	}
}
