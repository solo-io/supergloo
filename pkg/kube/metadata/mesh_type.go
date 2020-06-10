package metadata

import (
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
)

var (
	UnknownMeshType = func(mesh *discovery_v1alpha1.Mesh) error {
		return eris.Errorf("Unhandled mesh type in MeshToMeshType: %+v", mesh.Spec)
	}
)

func MeshToMeshType(mesh *discovery_v1alpha1.Mesh) (core_types.MeshType, error) {
	var meshType core_types.MeshType

	switch mesh.Spec.GetMeshType().(type) {
	case *discovery_types.MeshSpec_Istio1_5_:
		meshType = core_types.MeshType_ISTIO1_5
	case *discovery_types.MeshSpec_Istio1_6_:
		meshType = core_types.MeshType_ISTIO1_6
	case *discovery_types.MeshSpec_AwsAppMesh_:
		meshType = core_types.MeshType_APPMESH
	case *discovery_types.MeshSpec_ConsulConnect:
		meshType = core_types.MeshType_CONSUL
	case *discovery_types.MeshSpec_Linkerd:
		meshType = core_types.MeshType_LINKERD
	default:
		return meshType, UnknownMeshType(mesh)
	}

	return meshType, nil
}
