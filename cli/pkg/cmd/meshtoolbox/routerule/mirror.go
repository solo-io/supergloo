package routerule

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	glooV1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
)

func EnsureMirror(mOpts *options.InputMirror, opts *options.Options) error {

	var stagingUpstreams []*core.ResourceRef
	stagingUpstream := &core.ResourceRef{}

	if opts.Top.Static {
		upstreamClient, err := common.GetUpstreamClient()
		if err != nil {
			return err
		}
		var upstreams []*glooV1.Upstream
		if mOpts.Upstream != "" {
			upstreams, err = validateUpstreams(upstreamClient, mOpts.Upstream)
			if err != nil {
				return err
			}
			stagingUpstreams = toResourceRefs(upstreams)
			if len(stagingUpstreams) != 1 {
				return fmt.Errorf("Only one upstream spec is allowed for the mirror, received %v", len(stagingUpstreams))
			}
			stagingUpstream = stagingUpstreams[0]
		}
	} else {
		if err := nsutil.EnsureCommonResource("upstream", "Please select an upstream to use as a mirror.", stagingUpstream, opts); err != nil {
			return err
		}
	}
	opts.MeshTool.RoutingRule.Mirror = &glooV1.Destination{
		Upstream: *stagingUpstream,
	}
	return nil
}
