package table_printing

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
)

func DefaultTableBuilder() TableBuilder {
	return func(out io.Writer) *tablewriter.Table {
		table := tablewriter.NewWriter(out)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false) // ensure that newlines within a cell remain meaningful
		table.SetRowLine(true)       // print "-----" lines between rows
		return table
	}
}

func BuildSimpleStatusCell(status *zephyr_core_types.Status, statusName string) string {
	result := ""
	if status == nil {
		return result
	}

	result += fmt.Sprintf("%s Status:\n  State: %s", statusName, status.GetState())
	if status.GetState() != zephyr_core_types.Status_ACCEPTED {
		result += fmt.Sprintf("\n  Message: %s", status.GetMessage())
	}
	return result
}
