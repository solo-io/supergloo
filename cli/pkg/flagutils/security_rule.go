package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddCreateSecurityRuleFlags(set *pflag.FlagSet, in *options.CreateSecurityRule) {
	addSelectorFlags("source", "originating from", set, &in.SourceSelector)
	addSelectorFlags("dest", "sent to", set, &in.DestinationSelector)
	addTargetMeshFlags(set, &in.TargetMesh)
	addAllowedMethodsFlag(set, in)
	addAllowedPathsFlag(set, in)
}

func addAllowedMethodsFlag(set *pflag.FlagSet, in *options.CreateSecurityRule) {
	set.StringSliceVar(&in.AllowedMethods, "allowed-methods", nil, "HTTP methods that are allowed for this rule. "+
		"Leave empty to allow all paths")
}

func addAllowedPathsFlag(set *pflag.FlagSet, in *options.CreateSecurityRule) {
	set.StringSliceVar(&in.AllowedPaths, "allowed-paths", nil, "HTTP paths that are allowed for this rule. "+
		"Leave empty to allow all paths")
}
