package apply

import (
	"context"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/spf13/pflag"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var retryCommand = routingRuleSpecCommand{
	use:   "retries",
	alias: "rt",
	short: "apply a retry rule",
	long: `Retry rules are used to retry failed requests within a Mesh. 
The retries command contains subcommands for different types of retry policies. 
The retry policy you choose may only be compatible with a certain mesh type.
See documentation at https://supergloo.solo.io for more information.`,
	subCmds: retrySubCommands,
}

// source: https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-on
var possibleRetryOnValues = []string{
	"5xx",
	"gateway-error",
	"connect-failure",
	"retriable-4xx",
	"refused-stream",
	"retriable-status-codes",
	"cancelled",
	"deadline-exceeded",
	"internal",
	"resource-exhausted",
	"unavailable",
}

func isValidRetryOnValue(valueList string) bool {
	vals := strings.Split(valueList, ",")
	for _, val := range vals {
		var valid bool
		for _, possible := range possibleRetryOnValues {
			if val == possible {
				valid = true
				break
			}
		}
		if !valid {
			return false
		}
	}
	return true
}

var retrySubCommands = []routingRuleSpecCommand{
	{
		use:   "max",
		alias: "m",
		short: "apply a max retry policy rule. currently only used by managed Istio meshes",
		addFlagsFunc: func(set *pflag.FlagSet, in *options.RoutingRuleSpec) {
			flagutils.AddMaxRetriesFlags(set, &in.Retries.MaxRetries)
		},
		specSurveyFunc: func(ctx context.Context, in *options.CreateRoutingRule) error {
			return surveyutils.SurveyMaxRetries(&in.RoutingRuleSpec.Retries.MaxRetries)
		},
		convertSpecFunc: func(in options.RoutingRuleSpec) (*v1.RoutingRuleSpec, error) {
			if in.Retries.MaxRetries.Attempts == 0 {
				return nil, errors.Errorf("attempts cannot be 0")
			}
			if in.Retries.MaxRetries.PerTryTimeout < time.Second {
				return nil, errors.Errorf("per try timeout must be >=1s")
			}
			if !isValidRetryOnValue(in.Retries.MaxRetries.RetryOn) {
				return nil, errors.Errorf("invalid value for ")
			}
			return &v1.RoutingRuleSpec{
				RuleType: &v1.RoutingRuleSpec_Retries{
					Retries: &v1.RetryPolicy{
						MaxRetries: &v1alpha3.HTTPRetry{
							Attempts:      int32(in.Retries.MaxRetries.Attempts),
							PerTryTimeout: types.DurationProto(in.Retries.MaxRetries.PerTryTimeout),
							RetryOn:       in.Retries.MaxRetries.RetryOn,
						},
					},
				},
			}, nil
		},
		updateExistingFunc: func(old, new *v1.RoutingRuleSpec) {
			oldRetryPolicy := old.GetRetries()
			if oldRetryPolicy == nil {
				return
			}
			// preserve retry budget
			new.GetRetries().RetryBudget = oldRetryPolicy.RetryBudget
		},
	},
	{
		use:   "budget",
		alias: "b",
		short: "apply a retry budget policy. currently only used by managed Linkerd meshes",
		addFlagsFunc: func(set *pflag.FlagSet, in *options.RoutingRuleSpec) {
			flagutils.AddRetryBudgetFlags(set, &in.Retries.RetryBudget)
		},
		specSurveyFunc: func(ctx context.Context, in *options.CreateRoutingRule) error {
			return surveyutils.SurveyRetryBudget(&in.RoutingRuleSpec.Retries.RetryBudget)
		},
		convertSpecFunc: func(in options.RoutingRuleSpec) (*v1.RoutingRuleSpec, error) {
			if in.Retries.RetryBudget.RetryRatio < 0 || in.Retries.RetryBudget.RetryRatio > 1 {
				return nil, errors.Errorf("retry ratio must be a percentage between 0 and 1")
			}
			return &v1.RoutingRuleSpec{
				RuleType: &v1.RoutingRuleSpec_Retries{
					Retries: &v1.RetryPolicy{
						RetryBudget: &v1.RetryBudget{
							RetryRatio:          in.Retries.RetryBudget.RetryRatio,
							MinRetriesPerSecond: in.Retries.RetryBudget.MinRetriesPerSecond,
							Ttl:                 in.Retries.RetryBudget.Ttl,
						},
					},
				},
			}, nil
		},
		updateExistingFunc: func(old, new *v1.RoutingRuleSpec) {
			oldRetryPolicy := old.GetRetries()
			if oldRetryPolicy == nil {
				return
			}
			// preserve max retries
			new.GetRetries().MaxRetries = oldRetryPolicy.MaxRetries
		},
	},
}
