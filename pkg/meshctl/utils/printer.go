package utils

import (
	"fmt"
	"strings"

	"github.com/go-openapi/swag"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	Yaml = "yaml"
	Json = "json"
)

var validFormats = []string{Yaml, Json}

// augment the input cobra command with a flag specifying the output format, also add pre-run validation that the format string is valid
func AddOutputFlag(cmd *cobra.Command, format *string) {
	// augment existing PreRunE, if it exists, with validation of format flag
	existingPreRunE := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		formatVal := swag.StringValue(format)
		if !stringutils.ContainsString(formatVal, validFormats) {
			return eris.Errorf("invalid format %s. valid values are [%s]", formatVal, strings.Join(validFormats, ","))
		}

		if existingPreRunE == nil {
			return nil
		}
		return existingPreRunE(cmd, args)
	}

	// augment flags with format flag
	cmd.Flags().StringVarP(
		format,
		"output",
		"o",
		Yaml,
		fmt.Sprintf("set the output format, valid values are [%s]", strings.Join(validFormats, ",")),
	)
}

// marshal a proto message with the given format
func MarshalProtoWithFormat(msg proto.Message, format string) (string, error) {
	switch format {
	case Yaml:
		yaml, err := yaml.Marshal(msg)
		return string(yaml), err
	case Json:
		marshaler := &jsonpb.Marshaler{
			Indent: "  ",
		}
		return marshaler.MarshalToString(msg)
	default:
		return "", eris.Errorf("unrecognized output format %s", format)
	}
}
