package access_policy_enforcer

import (
	"github.com/rotisserie/eris"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
)

const (
	IstioDefaultAccessControlValue   = false
	AppmeshDefaultAccessControlValue = true
)

func DefaultAccessControlValueForMesh(mesh *zephyr_discovery.Mesh) (bool, error) {
	if mesh == nil {
		return false, nil
	}
	switch mesh.Spec.GetMeshType().(type) {
	case *discovery_types.MeshSpec_Istio:
		return IstioDefaultAccessControlValue, nil
	case *discovery_types.MeshSpec_AwsAppMesh_:
		return AppmeshDefaultAccessControlValue, nil
	case *discovery_types.MeshSpec_ConsulConnect:
		return false, nil
	case *discovery_types.MeshSpec_Linkerd:
		return false, nil
	default:
		return false, eris.Errorf("Unknown mesh type: %+v", mesh.Spec)
	}
}
