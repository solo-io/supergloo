package flagutils

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func addSelectorFlags(prefix, direction string, set *pflag.FlagSet, in *options.Selector) {
	set.Var(&in.SelectedUpstreams,
		prefix+"-"+"upstreams",
		fmt.Sprintf("apply this rule to requests %v these upstreams. format must be <NAMESPACE>.<NAME>.", direction))

	set.Var(&in.SelectedLabels,
		prefix+"-"+"labels",
		fmt.Sprintf("apply this rule to requests %v pods with these labels. format must be KEY=VALUE", direction))

	set.StringSliceVar(&in.SelectedNamespaces,
		prefix+"-"+"namespaces",
		nil,
		fmt.Sprintf("apply this rule to requests %v pods in these namespaces", direction))
}

func addTargetMeshFlags(set *pflag.FlagSet, in *options.ResourceRefValue) {
	set.Var(in,
		"target-mesh",
		"select the target mesh or mesh group to which to apply this rule. format must be NAMESPACE.NAME",
	)
}
