package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/pflag"
)

const (
	Description_MaxRetries_Attempts      = "REQUIRED. Number of retries for a given request. The interval between retries will be determined automatically (25ms+). Actual number of retries attempted depends on the httpReqTimeout."
	Description_MaxRetries_PerTryTimeout = "Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms. MUST BE >=1ms"
	Description_MaxRetries_RetryOn       = "Specifies the conditions under which retry takes place. One or more policies can be specified using a ‘,’ delimited list. The supported policies can be found in <https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-on> and <https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-grpc-on>"
)

func AddMaxRetriesFlags(set *pflag.FlagSet, opts *options.MaxRetries) {
	set.Uint32VarP(&opts.Attempts, "attempts", "a", 0, Description_MaxRetries_Attempts)
	set.DurationVarP(&opts.PerTryTimeout, "per-try-timeout", "t", 0, Description_MaxRetries_PerTryTimeout)
	set.StringVarP(&opts.RetryOn, "retry-on", "ro", "", Description_MaxRetries_RetryOn)
}

const (
	Description_RetryBudget_RetryRatio          = "the ratio of additional traffic that may be added by retries. retry_ratio of 0.1 means that 1 retry may be attempted for every 10 regular requests"
	Description_RetryBudget_MinRetriesPerSecond = "the proxy may always attempt this number of retries per second, even if it would violate the retryRatio"
	Description_RetryBudget_Ttl                 = "This duration indicates for how long requests should be considered for the purposes of enforcing the retryRatio.  A higher value considers a larger window and therefore allows burstier retries."
)

func AddRetryBudgetFlags(set *pflag.FlagSet, opts *v1.RetryBudget) {
	set.Float32VarP(&opts.RetryRatio, "ratio", "r", 0, Description_RetryBudget_RetryRatio)
	set.Uint32VarP(&opts.MinRetriesPerSecond, "min-retries", "m", 0, Description_RetryBudget_MinRetriesPerSecond)
	set.DurationVarP(&opts.Ttl, "ttl", "t", 0, Description_RetryBudget_Ttl)
}
