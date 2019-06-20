package surveyutils

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo/pkg/cliutil"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/supergloo/cli/pkg/options"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func SurveyTrafficShiftingSpec(ctx context.Context, in *options.CreateRoutingRule) error {
	ts := v1.TrafficShifting{
		Destinations: &gloov1.MultiDestination{
			Destinations: []*gloov1.WeightedDestination{},
		},
	}
	fmt.Println("select the upstreams to which you wish to direct traffic")
	ups, err := SurveyUpstreams(ctx)
	if err != nil {
		return err
	}
	for _, us := range ups {
		var weight uint32
		if err := cliutil.GetUint32Input(fmt.Sprintf("choose a weight for %v", us), &weight); err != nil {
			return err
		}

		ts.Destinations.Destinations = append(ts.Destinations.Destinations, &gloov1.WeightedDestination{
			Destination: &gloov1.Destination{
				DestinationType: &gloov1.Destination_Upstream{
					Upstream: &us,
				},
			},
			Weight: weight,
		})
	}

	in.RoutingRuleSpec.TrafficShifting = options.TrafficShiftingValue(ts)
	return nil
}
