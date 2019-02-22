package flagutils

import "github.com/spf13/pflag"

func AddOutputFlag(set *pflag.FlagSet, strptr *string) {
	set.StringVarP(strptr, "output", "o", "", "output format: (yaml, json, table)")
}

func AddInteractiveFlag(set *pflag.FlagSet, boolptr *bool) {
	set.BoolVarP(boolptr, "interactive", "i", false, "run in interactive mode")
}
