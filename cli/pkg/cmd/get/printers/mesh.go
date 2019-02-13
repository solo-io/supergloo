package printers

import (
	"io"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/solo-io/solo-kit/pkg/utils/cliutils"
	"github.com/solo-io/supergloo/cli/pkg/common"
	v1 "github.com/solo-io/supergloo/pkg2/api/v1"
)

func MeshTable(list *v1.MeshList, output string, template string) error {
	err := cliutils.PrintList(output, template, list,
		func(data interface{}, w io.Writer) error {
			meshTable(data.(*v1.MeshList), w)
			return nil
		},
		os.Stdout)
	return err
}

func meshTable(list *v1.MeshList, w io.Writer) {
	table := tablewriter.NewWriter(w)
	headers := []string{"", "name", "mesh-type", "namespace", "policy count", "encryption"}
	table.SetHeader(headers)

	table.SetBorder(false)

	for i, v := range *list {
		table.Append(transformMesh(v, i+1))
	}

	table.Render()
}

func transformMesh(mesh *v1.Mesh, index int) []string {
	var encryption string
	meshName := mesh.Metadata.Name
	target, namespace := getMeshType(mesh)
	policyCount := strconv.Itoa(getPolicyCount(mesh))
	if mesh.Encryption != nil {
		encryption = strconv.FormatBool(mesh.Encryption.TlsEnabled)
	} else {
		encryption = strconv.FormatBool(false)
	}
	row := []string{strconv.Itoa(index), meshName, target, namespace, policyCount, encryption}
	return row
}

func getPolicyCount(mesh *v1.Mesh) int {
	if mesh.Policy == nil {
		return 0
	}
	if mesh.Policy.Rules == nil {
		return 0
	}
	return len(mesh.Policy.Rules)
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
