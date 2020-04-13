package table_printing

import (
	"fmt"
	"io"
	"strings"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/internal"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

func NewAccessControlPolicyPrinter(tableBuilder TableBuilder) AccessControlPolicyPrinter {
	return &accessControlPolicyPrinter{
		tableBuilder: tableBuilder,
	}
}

type accessControlPolicyPrinter struct {
	tableBuilder TableBuilder
}

func (a *accessControlPolicyPrinter) Print(out io.Writer, printMode PrintMode, accessControlPolicies []*networking_v1alpha1.AccessControlPolicy) error {
	if len(accessControlPolicies) == 0 {
		return nil
	}

	table := a.tableBuilder(out)

	preFilteredHeaderRow := []string{
		"Name",
		"Source",
		"Destination",
		"Allowed Paths",
		"Allowed Methods",
		"Allowed Ports",
	}

	// we're always going to populate every column for every row
	// however, if a column is always empty, we'll filter out that column at the end before rendering
	rows := [][]string{}
	for _, acp := range accessControlPolicies {
		acSpec := acp.Spec

		newRow := []string{acp.GetName()}
		if acSpec.GetSourceSelector() != nil && printMode == ServicePrintMode {
			newRow = append(newRow, a.identitySelectorToCell(acSpec.SourceSelector))
		} else {
			newRow = append(newRow, "")
		}

		if acSpec.GetDestinationSelector() != nil && printMode == WorkloadPrintMode {
			newRow = append(newRow, internal.ServiceSelectorToCell(acSpec.DestinationSelector))
		} else {
			newRow = append(newRow, "")
		}

		if acSpec.GetAllowedPaths() != nil {
			newRow = append(newRow, a.allowedPathsToCell(acSpec.AllowedPaths))
		} else {
			newRow = append(newRow, "")
		}

		if acSpec.GetAllowedMethods() != nil {
			newRow = append(newRow, a.allowedMethodsToCell(acSpec.AllowedMethods))
		} else {
			newRow = append(newRow, "")
		}

		if acSpec.GetAllowedPorts() != nil {
			newRow = append(newRow, a.allowedPortsToCell(acSpec.AllowedPorts))
		} else {
			newRow = append(newRow, "")
		}

		rows = append(rows, newRow)
	}

	headersWithNonemptyColumns, rowsWithEmptyColumnsFiltered := internal.FilterEmptyColumns(preFilteredHeaderRow, rows)

	table.SetHeader(headersWithNonemptyColumns)
	table.AppendBulk(rowsWithEmptyColumnsFiltered)
	table.Render()
	return nil
}

func (a *accessControlPolicyPrinter) allowedPortsToCell(allowedPorts []uint32) string {
	portStrings := []string{}
	for _, port := range allowedPorts {
		portStrings = append(portStrings, fmt.Sprintf("%d", port))
	}
	return strings.Join(portStrings, "\n")
}

func (a *accessControlPolicyPrinter) allowedMethodsToCell(allowedMethods []core_types.HttpMethodValue) string {
	allowedMethodStrings := []string{}
	for _, method := range allowedMethods {
		allowedMethodStrings = append(allowedMethodStrings, method.String())
	}
	return strings.Join(allowedMethodStrings, "\n")
}

func (a *accessControlPolicyPrinter) allowedPathsToCell(allowedPaths []string) string {
	return strings.Join(allowedPaths, "\n")
}

func (a *accessControlPolicyPrinter) identitySelectorToCell(identitySelector *core_types.IdentitySelector) string {
	return "TODO" // TODO
}
