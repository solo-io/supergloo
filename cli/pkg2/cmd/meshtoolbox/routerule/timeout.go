package routerule

import (
	types "github.com/gogo/protobuf/types"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
)

func EnsureTimeout(opts *options.Options) error {
	opts.MeshTool.RoutingRule.Timeout = &types.Duration{}
	return EnsureDuration("Please specify the timeout", &(opts.Create.InputRoutingRule).Timeout, opts.MeshTool.RoutingRule.Timeout, opts)
}
