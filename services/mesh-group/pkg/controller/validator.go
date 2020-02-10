package controller

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery"
)

var (
	OnlyIstioSupportedError = func(meshName string) error {
		return eris.Errorf("Illegal mesh type found for group %s, currently only Istio is supported", meshName)
	}
	InvalidMeshRefsError = func(refs []string) error {
		return eris.Errorf("The following mesh refs are invalid: %v", refs)
	}
)

//go:generate mockgen -source ./validator.go -destination mocks/mock_validator.go
/*
	The MeshGroupValidator is meant to check the validity of any MeshGroup resource, and return the updated status for
	said resource. The properties it is testing are the ones which cannot be tested by a simple JSON schema check, as
	in the future this will be done using a schema in the CRD spec.
*/
type MeshGroupValidator interface {
	/*
		Validate takes as arguments the ctx of the event, as well as the mesh group needing to be validated.

		The return states are as follows:

		1. MeshGroupStatus_PROCESSING_ERROR, err: This means that there was an error trying to determine the
		validity of the mesh group, and the event should be requeued.
		2. MeshGroupStatus_INVALID, err: This means that the mesh group was processed properly, and there was
		found to be an error with the configuration. This event should not be requeued.
		3. MeshGroupStatus_VALID, nil: This means that the mesh group was processed properly, and the mesh group
		is valid
	*/
	Validate(ctx context.Context, mg *v1alpha1.MeshGroup) (types.MeshGroupStatus, error)
}

type meshGroupValidator struct {
	meshClient discovery.LocalMeshClient
}

func MeshGroupValidatorProvider(meshClient discovery.LocalMeshClient) MeshGroupValidator {
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

func getMeshesForRefs(refs []*types.ResourceRef, meshList *v1alpha1.MeshList) ([]*v1alpha1.Mesh, error) {
	var result []*v1alpha1.Mesh
	var invalidRefs []string
	for _, ref := range refs {
		var foundMesh *v1alpha1.Mesh
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
