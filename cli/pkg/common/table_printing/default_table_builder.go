package table_printing

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

var DefaultTableBuilder TableBuilder = func(out io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(out)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false) // ensure that newlines within a cell remain meaningful
	table.SetRowLine(true)       // print "-----" lines between rows
	return table
}
