package printers

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

const (
	minWidth     = 6
	width        = 4
	padding      = 4
	padCharacter = ' '
	flags        = 0
	tab          = "\t"
)

// Used to apply table formatting to the CLI output
type TableWriter struct {
	writer *tabwriter.Writer
}

// Returns a table writer that writes to stdout
func NewTableWriter(writer io.Writer) *TableWriter {

	return &TableWriter{
		writer: tabwriter.NewWriter(writer, minWidth, width, padding, padCharacter, flags),
	}
}

// Transforms the slice into properly formatted text and prints it to the writer
func (w *TableWriter) WriteLine(line []string) error {
	_, err := fmt.Fprintln(w.writer, strings.Join(line, tab))
	return err
}

func (w *TableWriter) Flush() error {
	return w.writer.Flush()
}
