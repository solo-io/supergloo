package status

import (
	"encoding/json"
	"fmt"
	"io"
)

type JsonPrinter StatusPrinter

func NewJsonPrinter() JsonPrinter {
	return &jsonPrinter{}
}

type jsonPrinter struct{}

func (p *jsonPrinter) Print(out io.Writer, statusReport *StatusReport) {
	result, _ := json.Marshal(statusReport)
	fmt.Fprintln(out, string(result))
}
