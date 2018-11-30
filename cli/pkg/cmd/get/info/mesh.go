package info

import (
	"strconv"

	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	meshName   = "NAME"
	namespace  = "NAMESPACE"
	target     = "TARGET-MESH"
	status     = "STATUS"
	encryption = "ENCRYPTION"
)

// TODO: ideally at some point we might annotate our .proto files with this information
var meshHeaders = []Header{
	{Name: meshName, WideOnly: false},
	{Name: namespace, WideOnly: false},
	{Name: target, WideOnly: false},
	{Name: status, WideOnly: false},
	{Name: encryption, WideOnly: true},
}

func FromMesh(mesh *v1.Mesh) *ResourceInfo {
	data := transformMesh(mesh)
	return &ResourceInfo{headers: meshHeaders, data: []map[string]string{data}}
}

func FromMeshList(list *v1.MeshList) *ResourceInfo {
	var data Data = make([]map[string]string, 0)
	for _, mesh := range *list {
		data = append(data, transformMesh(mesh))
	}
	return &ResourceInfo{headers: meshHeaders, data: data}
}

func transformMesh(mesh *v1.Mesh) map[string]string {
	var meshFieldMap = make(map[string]string, len(meshHeaders))
	meshFieldMap[meshName] = mesh.Metadata.Name
	meshFieldMap[target], meshFieldMap[namespace] = getMeshType(mesh)
	meshFieldMap[status] = core.Status_State_name[int32(mesh.Status.State)]
	if mesh.Encryption != nil {
		meshFieldMap[encryption] = strconv.FormatBool(mesh.Encryption.TlsEnabled)
	} else {
		meshFieldMap[encryption] = strconv.FormatBool(false)
	}
	return meshFieldMap
}

func getMeshType(mesh *v1.Mesh) (meshType, installationNamespace string) {
	if mesh.MeshType == nil {
		return "", ""
	}
	switch x := mesh.MeshType.(type) {
	case *v1.Mesh_Istio:
		return common.Istio, x.Istio.InstallationNamespace
	case *v1.Mesh_Consul:
		return common.Consul, x.Consul.InstallationNamespace
	case *v1.Mesh_Linkerd2:
		return common.Linkerd2, x.Linkerd2.InstallationNamespace
	default:
		//should never happen
		return "", ""
	}
}
