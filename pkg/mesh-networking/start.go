package mesh_networking

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/spf13/pflag"
)

type NetworkingOpts struct {
	*bootstrap.Options
	disallowIntersectingConfig bool
	watchOutputTypes           bool
}

func (opts *NetworkingOpts) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&opts.disallowIntersectingConfig, "disallow-intersecting-config", false, "if true, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh")
	flags.BoolVar(&opts.watchOutputTypes, "watch-output-types", true, "if true, Gloo Mesh will resync upon changes to the service mesh config output by Gloo Mesh")
}

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts *NetworkingOpts) error {
	return StartExtended(ctx, opts, func(_ bootstrap.StartParameters) ExtensionOpts {
		return ExtensionOpts{}
	}, false)
}
