package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddTrafficShiftingFlags(set *pflag.FlagSet, in *options.RoutingRuleSpec) {
	set.Var(&in.TrafficShifting, "destination", "append a traffic shifting destination. format must be "+
		"<NAMESPACE>.<NAME>:<WEIGHT>")
}
