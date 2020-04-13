package table_printing

import (
	"fmt"
	"io"
	"strings"

	types2 "github.com/gogo/protobuf/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/internal"
	security_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
)

func NewVirtualMeshMCSRPrinter(tableBuilder TableBuilder) VirtualMeshCSRPrinter {
	return &vmcsrPrinter{
		tableBuilder: tableBuilder,
	}
}

type vmcsrPrinter struct {
	tableBuilder TableBuilder
}

func (m *vmcsrPrinter) Print(
	out io.Writer,
	vmcsrs []*security_v1alpha1.VirtualMeshCertificateSigningRequest,
) error {

	if len(vmcsrs) == 0 {
		return nil
	}
	table := m.tableBuilder(out)

	commonHeaderRows := []string{
		"Metadata",
		"CSR",
		"Response",
	}
	var preFilteredRows [][]string
	for _, vmcsr := range vmcsrs {
		var newRow []string
		// Append commonn metadata
		newRow = append(newRow, m.buildMetadataCell(vmcsr))

		// Append mesehes
		newRow = append(newRow, m.buildCSRCell(vmcsr.Spec))

		// Append status data
		newRow = append(newRow, m.buildStatusCell(vmcsr.Status))

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(commonHeaderRows, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()
	return nil
}

func (m *vmcsrPrinter) buildMetadataCell(virtualMesh *security_v1alpha1.VirtualMeshCertificateSigningRequest) string {
	var items []string

	items = append(items,
		fmt.Sprintf("Name: %s", virtualMesh.GetName()),
		fmt.Sprintf("Namespace: %s", virtualMesh.GetNamespace()),
		"VirtualMesh:",
		fmt.Sprintf("  Name: %s", virtualMesh.Spec.GetVirtualMeshRef().GetName()),
		fmt.Sprintf("  Namespace: %s", virtualMesh.Spec.GetVirtualMeshRef().GetNamespace()),
	)

	return strings.Join(items, "\n")
}

func (m *vmcsrPrinter) buildCSRCell(spec types.VirtualMeshCertificateSigningRequestSpec) string {
	var items []string

	if spec.GetCertConfig() != nil {
		items = append(items,
			"Certificate Config:",
			fmt.Sprintf("  Hosts: %+v", spec.GetCertConfig().GetHosts()),
			fmt.Sprintf("  Org: %s", spec.GetCertConfig().GetOrg()),
			fmt.Sprintf("  Mesh Type: %s", strings.ToLower(spec.GetCertConfig().GetMeshType().String())),
		)
	}

	if len(spec.GetCsrData()) != 0 {
		items = append(items, "\nCSR Data: PRESENT (redacted)")
	} else {
		items = append(items, "\nCSR Data: NOT PRESENT ")
	}

	return strings.Join(items, "\n")
}

func (m *vmcsrPrinter) buildStatusCell(status types.VirtualMeshCertificateSigningRequestStatus) string {
	var items []string

	if federationCell := BuildSimpleStatusCell(
		status.GetComputedStatus(),
		"Issuer",
	); federationCell != "" {
		items = append(items, federationCell)
	} else {
		return ""
	}

	if status.GetThirdPartyApproval() != nil {
		items = append(items,
			"\nApproval Workflow:\n",
			fmt.Sprintf(
				"  Last Updated: %s",
				types2.TimestampString(status.GetThirdPartyApproval().GetLastUpdatedTime()),
			),
			fmt.Sprintf("  Status: %s", status.GetThirdPartyApproval().GetApprovalStatus()),
		)
	}

	if status.GetResponse() != nil {
		items = append(items,
			"\nCertificate Data:",
			"  Root Certificate: PRESENT (redacted)",
			"  CA Certificate: PRESENT (redacted)",
		)
	}
	return strings.Join(items, "\n")
}
