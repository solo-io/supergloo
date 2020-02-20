package group_controller

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
)

var (
	OnlyIstioSupportedError = func(meshName string) error {
		return eris.Errorf("Illegal mesh type found for group %s, currently only Istio is supported", meshName)
	}
	InvalidMeshRefsError = func(refs []string) error {
		return eris.Errorf("The following mesh refs are invalid: %v", refs)
	}
)

type meshGroupValidator struct {
	meshClient zephyr_core.MeshClient
}

func MeshGroupValidatorProvider(meshClient zephyr_core.MeshClient) MeshGroupValidator {
	return &meshGroupValidator{
		meshClient: meshClient,
	}
}

func (m *meshGroupValidator) Validate(ctx context.Context, mg *v1alpha1.MeshGroup) (types.MeshGroupStatus, error) {
	// TODO: Currently we are listing meshes from all namespaces, however, the namespace(s) should be configurable.
	meshList, err := m.meshClient.List(ctx)
	if err != nil {
		return types.MeshGroupStatus{
			Config: types.MeshGroupStatus_PROCESSING_ERROR,
		}, err
	}
	matchingMeshes, err := getMeshesForRefs(mg.Spec.GetMeshes(), meshList)
	if err != nil {
		return types.MeshGroupStatus{
			Config: types.MeshGroupStatus_INVALID,
		}, err
	}
	for _, v := range matchingMeshes {
		if v.Spec.GetIstio() == nil {
			return types.MeshGroupStatus{
				Config: types.MeshGroupStatus_INVALID,
			}, OnlyIstioSupportedError(v.GetName())
		}
	}
	return types.MeshGroupStatus{
		Config: types.MeshGroupStatus_VALID,
	}, nil
}

func getMeshesForRefs(refs []*core_types.ResourceRef,
	meshList *discoveryv1alpha1.MeshList) ([]*discoveryv1alpha1.Mesh, error) {

	var result []*discoveryv1alpha1.Mesh
	var invalidRefs []string
	for _, ref := range refs {
		var foundMesh *discoveryv1alpha1.Mesh
		for _, mesh := range meshList.Items {
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
