package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddFaultInjectionFlags(set *pflag.FlagSet, opts *options.FaultInjection) {
	set.Float64VarP(&opts.Percent, "percent", "p", 0, "percentage of traffic to fault-inject")
}

func AddFaultInjectionDelayFlags(set *pflag.FlagSet, opts *options.FaultInjectionDelay) {
	set.DurationVarP(&opts.Fixed, "delay", "d", 0, "delay to inject into requests")
}

func AddFaultInjectionAbortFlags(set *pflag.FlagSet, opts *options.FaultInjectionAbort) {
	set.Int32VarP(&opts.Http.HttpStatus, "status", "s", 404, "http status to abort requests with")
}
