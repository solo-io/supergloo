package info

import (
	"strconv"

	"github.com/solo-io/supergloo/cli/pkg/constants"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	name       = "NAME"
	namespace  = "NAMESPACE"
	targetMesh = "TARGET-MESH"
	status     = "STATUS"
	encryption = "ENCRYPTION"
)

type MeshInfo struct {
	header []Header
	data   Data
}

// TODO: ideally at some point we might annotate our .proto files with this information
var headers = []Header{
	{Name: name, WideOnly: false},
	{Name: namespace, WideOnly: false},
	{Name: targetMesh, WideOnly: false},
	{Name: status, WideOnly: false},
	{Name: encryption, WideOnly: true},
}

func (info MeshInfo) Headers(opts options.Get) []string {
	h := make([]string, 0)
	for _, header := range headers {
		// if this column is wideOnly, include it only if the "-o wide" option was supplied
		if !header.WideOnly || opts.Output == "wide" {
			h = append(h, header.String())
		}
	}
	return h
}

func (info MeshInfo) Resources(opts options.Get) [][]string {
	includedHeaders := info.Headers(opts)
	result := make([][]string, len(info.data))

	// for each mesh
	for i, meshFieldMap := range info.data {

		// for each of the columns that we want to display
		line := make([]string, len(includedHeaders))
		for j, h := range includedHeaders {
			val, ok := meshFieldMap[h]
			if !ok {
				val = ""
			}
			line[j] = val
		}
		result[i] = line
	}
	return result
}

func From(mesh *v1.Mesh) *MeshInfo {
	data := transform(mesh)
	return &MeshInfo{header: headers, data: []map[string]string{data}}
}

func FromList(list *v1.MeshList) *MeshInfo {
	var data Data = make([]map[string]string, 0)
	for _, mesh := range *list {
		data = append(data, transform(mesh))
	}
	return &MeshInfo{header: headers, data: data}
}

func transform(mesh *v1.Mesh) map[string]string {
	var meshFieldMap = make(map[string]string, 0)
	meshFieldMap[name] = mesh.Metadata.Name
	meshFieldMap[targetMesh], meshFieldMap[namespace] = getMeshType(mesh)
	meshFieldMap[status] = core.Status_State_name[int32(mesh.Status.State)]
	meshFieldMap[encryption] = strconv.FormatBool(mesh.Encryption.TlsEnabled)
	return meshFieldMap
}

func getMeshType(mesh *v1.Mesh) (meshType, installationNamespace string) {
	if mesh.MeshType == nil {
		return "", ""
	}
	switch x := mesh.MeshType.(type) {
	case *v1.Mesh_Istio:
		return constants.Istio, x.Istio.InstallationNamespace
	case *v1.Mesh_Consul:
		return constants.Consul, x.Consul.InstallationNamespace
	case *v1.Mesh_Linkerd2:
		return constants.Linkerd2, x.Linkerd2.InstallationNamespace
	default:
		//should never happen
		return "", ""
	}
}
