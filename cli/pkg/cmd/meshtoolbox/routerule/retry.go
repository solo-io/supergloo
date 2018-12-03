package routerule

import (
	"strconv"

	types "github.com/gogo/protobuf/types"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common/iutil"
	v1alpha3 "github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
)

func EnsureRetry(irOpts *options.InputRetry, opts *options.Options) error {
	// initialize the fields
	opts.MeshTool.RoutingRule.Retries = &v1alpha3.HTTPRetry{
		PerTryTimeout: &types.Duration{},
	}
	// Gather the attempt count
	if !opts.Top.Static {
		err := iutil.GetStringInput("Please specify the number of retry attempts", &irOpts.Attempts)
		if err != nil {
			return err
		}
	}
	// if not in interactive mode, timeout values will have already been passed
	if irOpts.Attempts != "" {
		att, err := strconv.Atoi(irOpts.Attempts)
		if err != nil {
			return err
		}
		opts.MeshTool.RoutingRule.Retries.Attempts = int32(att)
	}
	if err := EnsureDuration("Please specify the per-try timeout", &irOpts.PerTryTimeout,
		opts.MeshTool.RoutingRule.Retries.PerTryTimeout,
		opts); err != nil {
		return err
	}
	return nil
}
