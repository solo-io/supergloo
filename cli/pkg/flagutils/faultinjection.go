package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddFaultInjectionFlags(set *pflag.FlagSet, opts *options.RoutingRuleSpec) {
	set.Float64VarP(&opts.FaultInjection.Percent, "percent", "p", 0, "percentage of traffic to fault-inject")
}

func AddFaultInjectionDelayFlags(set *pflag.FlagSet, opts *options.RoutingRuleSpec) {
	set.DurationVarP(&opts.FaultInjection.Delay.Fixed, "duration", "d", 0, "duration of the delay")
}

func AddFaultInjectionAbortFlags(set *pflag.FlagSet, opts *options.RoutingRuleSpec) {
	set.Int32VarP(&opts.FaultInjection.Abort.Http.HttpStatus, "status", "s", 0, "http status to abort requests with")
}
