package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddUpstreamTlsFlags(set *pflag.FlagSet, in *options.EditUpstream) {
	set.StringVar(&in.MtlsMeshMetadata.Name, "mesh-name", "", "name for the mesh")
	set.StringVar(&in.MtlsMeshMetadata.Namespace, "mesh-namespace", "supergloo-system", "namespace for the mesh")
}
