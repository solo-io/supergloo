package flagutils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/spf13/pflag"
)

func AddOutputFlag(set *pflag.FlagSet, strptr *string) {
	set.StringVarP(strptr, "output", "o", "", "output format: (yaml, json, table)")
}

func AddInteractiveFlag(set *pflag.FlagSet, boolptr *bool) {
	set.BoolVarP(boolptr, "interactive", "i", false, "run in interactive mode")
}

func AddMetadataFlags(set *pflag.FlagSet, in *core.Metadata) {
	set.StringVar(&in.Name, "name", "", "name for the resource")
	set.StringVar(&in.Namespace, "namespace", "supergloo-system", "namespace for the resource")
}
