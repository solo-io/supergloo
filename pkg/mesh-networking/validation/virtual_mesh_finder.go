package vm_validation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

var (
	InvalidMeshRefsError = func(refs []string) error {
		return eris.Errorf("The following mesh refs are invalid: %v", refs)
	}
)

type virtualMeshFinder struct {
	meshClient smh_discovery.MeshClient
}

func NewVirtualMeshFinder(meshClient smh_discovery.MeshClient) VirtualMeshFinder {
	return &virtualMeshFinder{meshClient: meshClient}
}

func (v *virtualMeshFinder) GetMeshesForVirtualMesh(
	ctx context.Context,
	virtualMesh *smh_networking.VirtualMesh,
) ([]*smh_discovery.Mesh, error) {
	meshList, err := v.meshClient.ListMesh(ctx)
	if err != nil {
		return nil, err
	}
	var result []*smh_discovery.Mesh
	var invalidRefs []string
	for _, ref := range virtualMesh.Spec.GetMeshes() {
		var foundMesh *smh_discovery.Mesh
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
