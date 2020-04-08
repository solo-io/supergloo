package table_printing

import (
	"fmt"
	"io"
	"strings"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/internal"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
)

func NewMeshPrinter(tableBuilder TableBuilder) MeshPrinter {
	return &meshPrinter{
		tableBuilder: tableBuilder,
		commonHeaderRows: []string{
			"Name",
			"Cluster",
			"Namespace",
			"Version",
		},
	}
}

type meshPrinter struct {
	tableBuilder     TableBuilder
	commonHeaderRows []string
}

func (m *meshPrinter) Print(out io.Writer, meshes []*discovery_v1alpha1.Mesh) error {

	meshesByType := GetMeshByType(meshes)

	if err := m.printIstioMeshes(out, meshesByType.Istio); err != nil {
		return err
	}

	if err := m.printLinkerdMeshes(out, meshesByType.Linkerd); err != nil {
		return err
	}

	return nil
}

func (m *meshPrinter) printIstioMeshes(out io.Writer, meshes []*discovery_v1alpha1.Mesh) error {
	// Do nothing if no meshes exist
	if len(meshes) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nIstio:")

	table := m.tableBuilder(out)

	istioHeaderRows := []string{
		"CitadelInfo",
	}
	commonHeaderRows := append(m.commonHeaderRows, istioHeaderRows...)

	var preFilteredRows [][]string
	for _, mesh := range meshes {
		// Append commonn metadata
		newRow := []string{mesh.GetName(), mesh.Spec.GetCluster().GetName()}

		// istioMeshes should be prefiltered, if they are not, simply skip it
		istioMesh := mesh.Spec.GetIstio()
		if istioMesh == nil {
			continue
		}
		// Append installation info
		newRow = append(
			newRow,
			istioMesh.GetInstallation().GetInstallationNamespace(),
			istioMesh.GetInstallation().GetVersion(),
		)

		// Append Citadel Info
		if istioMesh.GetCitadelInfo() != nil {
			newRow = append(newRow, m.buildCitadelInfo(istioMesh.GetCitadelInfo()))
		}

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()
	return nil
}

func (m *meshPrinter) buildCitadelInfo(info *types.MeshSpec_IstioMesh_CitadelInfo) string {
	lines := []string{
		fmt.Sprintf("Trust Domain: %s", info.GetTrustDomain()),
		fmt.Sprintf("Citadel Namespace: %s", info.GetCitadelNamespace()),
		fmt.Sprintf("Citadel Service Account: %s", info.GetCitadelServiceAccount()),
	}
	return strings.Join(lines, "\n")
}

func (m *meshPrinter) printLinkerdMeshes(out io.Writer, meshes []*discovery_v1alpha1.Mesh) error {
	// Do nothing if no meshes exist
	if len(meshes) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nLinkerd:")

	table := m.tableBuilder(out)

	commonHeaderRows := append([]string{}, m.commonHeaderRows...)

	var preFilteredRows [][]string
	for _, mesh := range meshes {
		// Append commonn metadata
		newRow := []string{mesh.GetName(), mesh.Spec.GetCluster().GetName()}

		// linkerdMesh should be prefiltered, if they are not, simply skip it
		linkerdMesh := mesh.Spec.GetLinkerd()
		if linkerdMesh == nil {
			continue
		}
		// Append installation info
		newRow = append(
			newRow,
			linkerdMesh.GetInstallation().GetInstallationNamespace(),
			linkerdMesh.GetInstallation().GetVersion(),
		)

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()

	return nil
}

func (m *meshPrinter) printConsulConnectMeshes(out io.Writer, meshes []*discovery_v1alpha1.Mesh) error {
	// Do nothing if no meshes exist
	if len(meshes) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nConsul Connect:")

	table := m.tableBuilder(out)

	commonHeaderRows := append([]string{}, m.commonHeaderRows...)

	var preFilteredRows [][]string
	for _, mesh := range meshes {
		// Append commonn metadata
		newRow := []string{mesh.GetName(), mesh.Spec.GetCluster().GetName()}

		// Consul Connect should be prefiltered, if they are not, simply skip it
		consulConnectmesh := mesh.Spec.GetConsulConnect()
		if consulConnectmesh == nil {
			continue
		}
		// Append installation info
		newRow = append(
			newRow,
			consulConnectmesh.GetInstallation().GetInstallationNamespace(),
			consulConnectmesh.GetInstallation().GetVersion(),
		)

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()

	return nil
}

func (m *meshPrinter) printAwsAppMeshMeshes(out io.Writer, meshes []*discovery_v1alpha1.Mesh) error {
	// Do nothing if no meshes exist
	if len(meshes) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nAws AppMesh:")

	table := m.tableBuilder(out)

	commonHeaderRows := append([]string{}, m.commonHeaderRows...)

	var preFilteredRows [][]string
	for _, mesh := range meshes {
		// Append commonn metadata
		newRow := []string{mesh.GetName(), mesh.Spec.GetCluster().GetName()}

		// aws appmesh should be prefiltered, if they are not, simply skip it
		awsAppMesh := mesh.Spec.GetAwsAppMesh()
		if awsAppMesh == nil {
			continue
		}
		// Append installation info
		newRow = append(
			newRow,
			awsAppMesh.GetInstallation().GetInstallationNamespace(),
			awsAppMesh.GetInstallation().GetVersion(),
		)

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()

	return nil
}

type MeshByType struct {
	Istio         []*discovery_v1alpha1.Mesh
	Linkerd       []*discovery_v1alpha1.Mesh
	ConsulConnect []*discovery_v1alpha1.Mesh
	AwsAppMesh    []*discovery_v1alpha1.Mesh
}

func GetMeshByType(meshes []*discovery_v1alpha1.Mesh) *MeshByType {
	mbt := &MeshByType{}
	for _, mesh := range meshes {
		switch mesh.Spec.GetMeshType().(type) {
		case *types.MeshSpec_Istio:
			mbt.Istio = append(mbt.Istio, mesh)
		case *types.MeshSpec_Linkerd:
			mbt.Linkerd = append(mbt.Linkerd, mesh)
		case *types.MeshSpec_AwsAppMesh_:
			mbt.AwsAppMesh = append(mbt.AwsAppMesh, mesh)
		case *types.MeshSpec_ConsulConnect:
			mbt.ConsulConnect = append(mbt.ConsulConnect, mesh)
		}
	}
	return mbt
}
