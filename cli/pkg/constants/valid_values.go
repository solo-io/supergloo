package constants

// source: https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-on
var PossibleMaxRetry_RetryOnValues = []string{
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
