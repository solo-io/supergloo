package info

import (
	"strings"

	"github.com/solo-io/supergloo/pkg/api/v1"
)

type MeshInfo struct {
	Header []Header
	Data   Data
}

var headers = []Header{
	{Name: "name", WideOnly: false},
	{Name: "namespace", WideOnly: false},
	{Name: "target-mesh", WideOnly: false},
	{Name: "status", WideOnly: false},
	{Name: "encryption", WideOnly: true},
}

func (info MeshInfo) Headers() string {

	headers := make([]string, 0)
	for _, header := range info.Header {
		headers = append(headers, header.String())
	}

	// TODO: proper separation with padding
	return strings.Join(headers, "\t")
}

func (info MeshInfo) Resources() []string {
	// TODO implement
	return []string{"dummy"}
}

func From(mesh *v1.Mesh) *MeshInfo {
	data := Transform(mesh)
	return &MeshInfo{Header: headers, Data: []map[string]Field{data}}
}

func FromList(list *v1.MeshList) *MeshInfo {
	var data Data = make([]map[string]Field, 0)
	for _, mesh := range *list {
		data = append(data, Transform(mesh))
	}
	return &MeshInfo{Header: headers, Data: data}
}

// Transform
func Transform(mesh *v1.Mesh) map[string]Field {
	// TODO: implement transform
	return make(map[string]Field, 0)
}
