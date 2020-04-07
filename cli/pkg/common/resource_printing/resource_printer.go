package resource_printing

import (
	"io"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

var (
	ValidFormats               = []string{JSONFormat, YAMLFormat}
	ResourcePrinterProviderSet = wire.NewSet(
		NewResourcePrinter,
	)
	InvalidOutputFormatError = func(format string) error {
		return eris.Errorf("Invalid output format: %s. Must be one of (%s)", format, strings.Join([]string{JSONFormat, YAMLFormat}, "|"))
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

func (r *resourcePrinters) Print(out io.Writer, object runtime.Object, format string) error {
	switch format {
	case JSONFormat:
		return r.jsonPrinter.PrintObj(object, out)
	case YAMLFormat:
		return r.yamlPrinter.PrintObj(object, out)
	default:
		return InvalidOutputFormatError(format)
	}
}
