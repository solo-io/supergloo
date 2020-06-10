package table_printing

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/internal"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
)

func NewMeshServicePrinter(tableBuilder TableBuilder) MeshServicePrinter {
	return &meshServicePrinter{
		tableBuilder: tableBuilder,
	}
}

type meshServicePrinter struct {
	tableBuilder TableBuilder
}

func (m *meshServicePrinter) Print(
	out io.Writer,
	meshServices []*smh_discovery.MeshService,
) error {
	if len(meshServices) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nKubernetes Mesh Services:")

	table := m.tableBuilder(out)

	commonHeaderRows := []string{
		"Mesh",
		"Service + Ports",
		"Labels + Subsets",
		"Federation Data",
		"Status",
	}
	var preFilteredRows [][]string
	for _, meshService := range meshServices {
		var newRow []string
		// Append common metadata
		newRow = append(newRow, meshService.Spec.GetMesh().GetName())

		kubeService := meshService.Spec.GetKubeService()
		// Append Service data + Ports
		newRow = append(newRow, m.buildServiceDataCell(kubeService))

		// Append Labels
		newRow = append(newRow, m.buildLabelCell(meshService))

		// Append federation data
		newRow = append(newRow, m.buildFederationDataCell(meshService.Spec.GetFederation()))

		// Append status data
		newRow = append(newRow, m.buildStatusCell(meshService.Status))

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()
	return nil
}

func (m *meshServicePrinter) buildLabelCell(meshSvc *smh_discovery.MeshService) string {
	var items []string

	svc := meshSvc.Spec.GetKubeService()

	if len(svc.GetLabels()) != 0 {
		var labels []string
		for k, v := range svc.GetLabels() {
			labels = append(labels, fmt.Sprintf("  %s: %s", k, v))
		}
		sort.Strings(labels)
		items = append(items, fmt.Sprintf("Service Labels:\n%s", strings.Join(labels, "\n\n")))
	}

	if len(svc.GetWorkloadSelectorLabels()) != 0 {
		var labels []string
		for k, v := range svc.GetWorkloadSelectorLabels() {
			labels = append(labels, fmt.Sprintf("  %s: %s", k, v))
		}
		sort.Strings(labels)
		items = append(items, fmt.Sprintf("Workload Selector Labels:\n%s\n", strings.Join(labels, "\n")))
	}

	// Append subsets
	if len(meshSvc.Spec.GetSubsets()) != 0 {
		items = append(items, "Subsets:")
		var subsets []string
		for key, subset := range meshSvc.Spec.GetSubsets() {
			var subsetValues []string
			for _, subsetValue := range subset.Values {
				subsetValues = append(subsetValues, fmt.Sprintf("  - %s", subsetValue))
			}
			subsets = append(subsets, fmt.Sprintf("  %s:\n%s", key, strings.Join(subsetValues, "\n")))
		}
		sort.Strings(subsets)
		items = append(items, strings.Join(subsets, "\n"))
	}

	return strings.Join(items, "\n")
}

func (m *meshServicePrinter) buildServiceDataCell(svc *discovery_types.MeshServiceSpec_KubeService) string {
	if svc == nil {
		return ""
	}

	var items []string

	items = append(items, fmt.Sprintf(
		"Name: %s\nNamespace: %s\nCluster: %s\n",
		svc.GetRef().GetName(),
		svc.GetRef().GetNamespace(),
		svc.GetRef().GetCluster(),
	))

	if len(svc.GetPorts()) != 0 {
		items = append(items, "Ports:")
		for _, port := range svc.GetPorts() {
			items = append(items, fmt.Sprintf(
				"- Name: %s\n  Port: %d\n  Protocol: %s",
				port.GetName(),
				port.GetPort(),
				port.GetProtocol(),
			))
		}
	}

	return strings.Join(items, "\n")
}

func (m *meshServicePrinter) buildFederationDataCell(federation *discovery_types.MeshServiceSpec_Federation) string {
	if federation == nil {
		return ""
	}
	var items []string
	items = append(items, fmt.Sprintf("Multi Cluster DNS Name: %s", federation.GetMulticlusterDnsName()))

	if len(federation.GetFederatedToWorkloads()) != 0 {
		items = append(items, "Accessible Via:")
		for _, v := range federation.GetFederatedToWorkloads() {
			items = append(items, fmt.Sprintf("  - %s", v.GetName()))
		}
	}

	return strings.Join(items, "\n")
}

func (m *meshServicePrinter) buildStatusCell(status discovery_types.MeshServiceStatus) string {
	var items []string

	if federationCell := BuildSimpleStatusCell(status.GetFederationStatus(), "Federation"); federationCell != "" {
		items = append(items, federationCell)
	}

	return strings.Join(items, "\n")
}
