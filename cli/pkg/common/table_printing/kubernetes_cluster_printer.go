package table_printing

import (
	"fmt"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/internal"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

func NewKubernetesClusterPrinter(tableBuilder TableBuilder) KubernetesClusterPrinter {
	return &kubernetesClusterPrinter{
		tableBuilder: tableBuilder,
	}
}

type kubernetesClusterPrinter struct {
	tableBuilder TableBuilder
}

func (m *kubernetesClusterPrinter) Print(out io.Writer, clusters []*discovery_v1alpha1.KubernetesCluster) error {
	if len(clusters) == 0 {
		return nil
	}
	fmt.Fprintln(out, "\nKubernetes Clusters:")

	table := m.tableBuilder(out)

	commonHeaderRows := []string{
		"Name",
		"Version",
		"Cloud",
		"Write Namespace",
		"Secret Ref",
	}

	var preFilteredRows [][]string
	for _, cluster := range clusters {
		// Append common string data
		newRow := []string{
			cluster.GetName(),
			cluster.Spec.GetVersion(),
			cluster.Spec.GetCloud(),
			cluster.Spec.GetWriteNamespace(),
		}

		newRow = append(newRow,
			fmt.Sprintf(
				"%s.%s",
				cluster.Spec.GetSecretRef().GetName(),
				cluster.Spec.GetSecretRef().GetNamespace(),
			),
		)

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()
	return nil
}
