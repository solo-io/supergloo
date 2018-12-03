package routerule

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/cli/pkg/common/iutil"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	glooV1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func EnsureTrafficShifting(irOpts *options.InputTrafficShifting, opts *options.Options) error {
	// initialize the fields
	opts.MeshTool.RoutingRule.TrafficShifting = &superglooV1.TrafficShifting{
		Destinations: []*superglooV1.WeightedDestination{},
	}
	tsOpts := &(opts.Create.InputRoutingRule).TrafficShifting

	var stagingUpstreams []*core.ResourceRef
	var stagingWeights []uint32

	if opts.Top.Static {
		upstreamClient, err := common.GetUpstreamClient()
		if err != nil {
			return err
		}
		var upstreams []*glooV1.Upstream
		if tsOpts.Upstreams != "" {
			upstreams, err = validateUpstreams(upstreamClient, tsOpts.Upstreams)
			if err != nil {
				return err
			}
			stagingUpstreams = toResourceRefs(upstreams)
			if tsOpts.Weights == "" {
				return fmt.Errorf("You must provide a list of weights to apply traffic sharing rules.")
			}
			if err := getWeightsFromCsv(&stagingWeights, tsOpts.Weights, len(stagingUpstreams)); err != nil {
				return err
			}
		}
	} else {
		if err := nsutil.EnsureCommonResources("upstream", "Please select upstream(s) to recieve traffic", &stagingUpstreams, opts); err != nil {
			return err
		}
		var weightInput string
		if err := iutil.GetStringInput("Please specify the associated upstream's traffic sharing percentage as comma-separated integers (ex: 20,30,50).", &weightInput); err != nil {
			return err
		}
		if err := getWeightsFromCsv(&stagingWeights, weightInput, len(stagingUpstreams)); err != nil {
			return err
		}
	}
	return attachTrafficShiftingSpec(&(opts.MeshTool).RoutingRule, stagingUpstreams, stagingWeights)
}

func attachTrafficShiftingSpec(rr *superglooV1.RoutingRule, us []*core.ResourceRef, weights []uint32) error {
	// we should have already checked this on entry, just to be safe:
	if len(us) != len(weights) {
		return fmt.Errorf("Invalid traffic sharing specification, please provide one weight spec per upstream")
	}
	destinations := []*superglooV1.WeightedDestination{}
	for i, u := range us {
		destinations = append(destinations, &superglooV1.WeightedDestination{
			Upstream: u,
			Weight:   weights[i],
		})
	}
	rr.TrafficShifting = &superglooV1.TrafficShifting{
		Destinations: destinations,
	}
	return nil
}

func getWeightsFromCsv(list *[]uint32, csv string, expectedLen int) error {
	parts := strings.Split(csv, ",")
	if len(parts) != expectedLen {
		return fmt.Errorf("Must pass one weight weight per upstream. Received %v, expected %v", len(parts), expectedLen)
	}
	for _, s := range parts {
		n, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*list = append(*list, uint32(n))
	}
	return nil
}
