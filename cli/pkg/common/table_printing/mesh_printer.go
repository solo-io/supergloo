package table_printing

import (
	"fmt"
	"io"
	"strings"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/internal"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
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

func (m *meshPrinter) Print(out io.Writer, meshes []*zephyr_discovery.Mesh) error {

	meshesByType := getMeshByType(meshes)

	if err := m.printIstioMeshes(out, meshesByType.istio1_5); err != nil {
		return err
	}

	if err := m.printIstioMeshes(out, meshesByType.istio1_6); err != nil {
		return err
	}

	if err := m.printLinkerdMeshes(out, meshesByType.linkerd); err != nil {
		return err
	}

	if err := m.printAwsAppMeshMeshes(out, meshesByType.awsAppMesh); err != nil {
		return err
	}

	return nil
}

func (m *meshPrinter) printIstioMeshes(out io.Writer, meshes []*istioMesh) error {
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
		// Append common metadata
		newRow := []string{mesh.meshName, mesh.clusterName}

		// Append installation info
		newRow = append(
			newRow,
			mesh.istioMetadata.GetInstallation().GetInstallationNamespace(),
			mesh.istioMetadata.GetInstallation().GetVersion(),
		)

		// Append Citadel Info
		if mesh.istioMetadata.GetCitadelInfo() != nil {
			newRow = append(newRow, m.buildCitadelInfo(mesh.istioMetadata.GetCitadelInfo()))
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

func (m *meshPrinter) printLinkerdMeshes(out io.Writer, meshes []*zephyr_discovery.Mesh) error {
	// Do nothing if no meshes exist
	if len(meshes) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nLinkerd:")

	table := m.tableBuilder(out)

	commonHeaderRows := append([]string{}, m.commonHeaderRows...)

	var preFilteredRows [][]string
	for _, mesh := range meshes {
		// Append common metadata
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

func (m *meshPrinter) printConsulConnectMeshes(out io.Writer, meshes []*zephyr_discovery.Mesh) error {
	// Do nothing if no meshes exist
	if len(meshes) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nConsul Connect:")

	table := m.tableBuilder(out)

	commonHeaderRows := append([]string{}, m.commonHeaderRows...)

	var preFilteredRows [][]string
	for _, mesh := range meshes {
		// Append common metadata
		newRow := []string{mesh.GetName(), mesh.Spec.GetCluster().GetName()}

		// Consul Connect should be prefiltered, if they are not, simply skip it
		consulConnectMesh := mesh.Spec.GetConsulConnect()
		if consulConnectMesh == nil {
			continue
		}
		// Append installation info
		newRow = append(
			newRow,
			consulConnectMesh.GetInstallation().GetInstallationNamespace(),
			consulConnectMesh.GetInstallation().GetVersion(),
		)

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()

	return nil
}

func (m *meshPrinter) printAwsAppMeshMeshes(out io.Writer, meshes []*zephyr_discovery.Mesh) error {
	// Do nothing if no meshes exist
	if len(meshes) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nAws AppMesh:")

	table := m.tableBuilder(out)

	commonHeaderRows := append([]string{}, m.commonHeaderRows...)

	var preFilteredRows [][]string
	for _, mesh := range meshes {
		// Append common metadata
		newRow := []string{mesh.GetName(), mesh.Spec.GetCluster().GetName()}

		// aws appmesh should be prefiltered, if they are not, simply skip it
		awsAppMesh := mesh.Spec.GetAwsAppMesh()
		if awsAppMesh == nil {
			continue
		}
		// Append AppMesh instance info
		newRow = append(
			newRow,
			awsAppMesh.GetName(),
			awsAppMesh.GetAwsAccountId(),
			awsAppMesh.GetRegion(),
		)

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()

	return nil
}

type istioMesh struct {
	meshName      string
	clusterName   string
	istioMetadata *types.MeshSpec_IstioMesh
}

type meshByType struct {
	istio1_5      []*istioMesh
	istio1_6      []*istioMesh
	linkerd       []*zephyr_discovery.Mesh
	consulConnect []*zephyr_discovery.Mesh
	awsAppMesh    []*zephyr_discovery.Mesh
}

func getMeshByType(meshes []*zephyr_discovery.Mesh) *meshByType {
	mbt := &meshByType{}
	for _, mesh := range meshes {
		switch mesh.Spec.GetMeshType().(type) {
		case *types.MeshSpec_Istio1_5_:
			mbt.istio1_5 = append(mbt.istio1_5, &istioMesh{
				meshName:      mesh.GetName(),
				clusterName:   mesh.Spec.GetCluster().GetName(),
				istioMetadata: mesh.Spec.GetIstio1_5().GetMetadata(),
			})
		case *types.MeshSpec_Istio1_6_:
			mbt.istio1_6 = append(mbt.istio1_6, &istioMesh{
				meshName:      mesh.GetName(),
				clusterName:   mesh.Spec.GetCluster().GetName(),
				istioMetadata: mesh.Spec.GetIstio1_6().GetMetadata(),
			})
		case *types.MeshSpec_Linkerd:
			mbt.linkerd = append(mbt.linkerd, mesh)
		case *types.MeshSpec_AwsAppMesh_:
			mbt.awsAppMesh = append(mbt.awsAppMesh, mesh)
		case *types.MeshSpec_ConsulConnect:
			mbt.consulConnect = append(mbt.consulConnect, mesh)
		}
	}
	return mbt
}
