package vm_validation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
)

var (
	InvalidMeshRefsError = func(refs []string) error {
		return eris.Errorf("The following mesh refs are invalid: %v", refs)
	}
)

type virtualMeshFinder struct {
	meshClient zephyr_discovery.MeshClient
}

func NewVirtualMeshFinder(meshClient zephyr_discovery.MeshClient) VirtualMeshFinder {
	return &virtualMeshFinder{meshClient: meshClient}
}

func (g *virtualMeshFinder) GetMeshesForVirtualMesh(
	ctx context.Context,
	virtualMesh *v1alpha1.VirtualMesh,
) ([]*discoveryv1alpha1.Mesh, error) {
	meshList, err := g.meshClient.List(ctx)
	if err != nil {
		return nil, err
	}
	var result []*discoveryv1alpha1.Mesh
	var invalidRefs []string
	for _, ref := range virtualMesh.Spec.GetMeshes() {
		var foundMesh *discoveryv1alpha1.Mesh
		for _, mesh := range meshList.Items {
			// thankx rob pike
			mesh := mesh
			if mesh.GetName() == ref.GetName() && mesh.GetNamespace() == ref.GetNamespace() {
				foundMesh = &mesh
			}
		}
		if foundMesh == nil {
			invalidRefs = append(invalidRefs, fmt.Sprintf("%s.%s", ref.GetName(), ref.GetNamespace()))
			continue
		}
		result = append(result, foundMesh)
	}
	if len(invalidRefs) > 0 {
		return nil, InvalidMeshRefsError(invalidRefs)
	}
	return result, nil
}
