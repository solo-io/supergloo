package resource_printing

import (
	"io"
	"strings"

	"github.com/rotisserie/eris"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

type OutputFormat string

func (o OutputFormat) String() string {
	return string(o)
}

const (
	JSONFormat OutputFormat = "json"
	YAMLFormat OutputFormat = "yaml"
)

var (
	ValidFormats             = []string{JSONFormat.String(), YAMLFormat.String()}
	InvalidOutputFormatError = func(format OutputFormat) error {
		return eris.Errorf(
			"Invalid output format: %s. Must be one of (%s)",
			format.String(),
			strings.Join(ValidFormats, "|"))
	}
)

type resourcePrinters struct {
	jsonPrinter *printers.JSONPrinter
	yamlPrinter *printers.YAMLPrinter
}

func NewResourcePrinter() ResourcePrinter {
	return &resourcePrinters{
		jsonPrinter: &printers.JSONPrinter{},
		yamlPrinter: &printers.YAMLPrinter{},
	}
}

func (r *resourcePrinters) Print(out io.Writer, object runtime.Object, format OutputFormat) error {
	switch format {
	case JSONFormat:
		return r.jsonPrinter.PrintObj(object, out)
	case YAMLFormat:
		return r.yamlPrinter.PrintObj(object, out)
	default:
		return InvalidOutputFormatError(format)
	}
}
