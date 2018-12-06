package cliutils

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
)

// Printer represents a function that prints a value to io.Writer, usually using
// a table
type Printer func(interface{}, io.Writer) error

// Print - prints the given proto.Message to io.Writer using the specified output format
func Print(output, template string, m proto.Message, tblPrn Printer, w io.Writer) error {
	switch strings.ToLower(output) {
	case "yaml":
		return PrintYAML(m, w)
	case "yml":
		return PrintYAML(m, w)
	case "json":
		return PrintJSON(m, w)
	case "template":
		return PrintTemplate(m, template, w)
	default:
		return tblPrn(m, w)
	}
}

// PrintJSON - prints the given proto.Message to io.Writer in JSON
func PrintJSON(m proto.Message, w io.Writer) error {
	b, err := protoutils.MarshalBytes(m)
	if err != nil {
		return errors.Wrap(err, "unable to convert to JSON")
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

// PrintYAML - prints the given proto.Message to io.Writer in YAML
func PrintYAML(m proto.Message, w io.Writer) error {
	jsn, err := protoutils.MarshalBytes(m)
	if err != nil {
		return errors.Wrap(err, "unable to marshal")
	}
	b, err := yaml.JSONToYAML(jsn)
	if err != nil {
		return errors.Wrap(err, "unable to convert to YAML")
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

// PrintTemplate prints the give value using the provided Go template to io.Writer
func PrintTemplate(data interface{}, tmpl string, w io.Writer) error {
	t, err := template.New("output").Parse(tmpl)
	if err != nil {
		return errors.Wrap(err, "unable to parse template")
	}
	return t.Execute(w, data)
}

// PrintList - prints the given list of values to io.Writer using the specified output format
func PrintList(output, template string, list interface{}, tblPrn Printer, w io.Writer) error {
	switch strings.ToLower(output) {
	case "yaml":
		return PrintYAMLList(list, w)
	case "yml":
		return PrintYAMLList(list, w)
	case "json":
		return PrintJSONList(list, w)
	case "template":
		return PrintTemplate(list, template, w)
	default:
		return tblPrn(list, w)
	}
}

// PrintJSONList - prints the given list to io.Writer in JSON
func PrintJSONList(data interface{}, w io.Writer) error {
	list := reflect.ValueOf(data)
	_, err := fmt.Fprintln(w, "[")
	if err != nil {
		return errors.Wrap(err, "unable to print JSON list")
	}
	for i := 0; i < list.Len(); i++ {
		v, ok := list.Index(i).Interface().(proto.Message)
		if !ok {
			return errors.New("unable to convert to proto message")
		}
		if i != 0 {
			_, err = fmt.Fprintln(w, ",")
			if err != nil {
				return errors.Wrap(err, "unable to print JSON list")
			}
		}
		err = PrintJSON(v, w)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w, "]")
	return err
}

// PrintYAMLList - prints the given list to io.Writer in YAML
func PrintYAMLList(data interface{}, w io.Writer) error {
	list := reflect.ValueOf(data)
	for i := 0; i < list.Len(); i++ {
		v, ok := list.Index(i).Interface().(proto.Message)
		if !ok {
			return errors.New("unable to convert to proto message")
		}
		if _, err := fmt.Fprintln(w, "---"); err != nil {
			return errors.Wrap(err, "unable to print YAML list")
		}
		if err := PrintYAML(v, w); err != nil {
			return err
		}
	}
	return nil
}
