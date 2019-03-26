package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddUpstreamTlsFlags(set *pflag.FlagSet, in *options.EditUpstream) {
	set.VarP(&in.MtlsMesh, "target-mesh", "t", "target mesh")
}
