package routerule

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func AddBaseFlags(cmd *cobra.Command, opts *options.Options) {
	rrOpts := &(opts.Create).InputRoutingRule
	flags := cmd.Flags()

	flags.StringVar(&(rrOpts.TargetMesh).Name,
		"mesh",
		"",
		"The mesh that will be the target for this rule")

	flags.StringVarP(&(rrOpts.TargetMesh).Namespace,
		"namespace", "n",
		"",
		"The namespace for this routing rule. Defaults to \"default\"")

	flags.StringVar(&rrOpts.Sources,
		"sources",
		"",
		"Sources for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon.")

	flags.StringVar(&rrOpts.Destinations,
		"destinations",
		"",
		"Destinations for this rule. Same format as for 'sources'")

	flags.BoolVar(&rrOpts.OverrideExisting,
		"override",
		false,
		"If set to \"true\", the command will override any existing routing rule that matches the given namespace and name")

	rrOpts.Matchers = *flags.StringArrayP("matchers",
		"m",
		nil,
		"Matcher for this rule")
}

func AddTimeoutFlags(cmd *cobra.Command, opts *options.Options) {
	rrOpts := &(opts.Create).InputRoutingRule
	flags := cmd.Flags()
	flags.StringVar(&(rrOpts.Timeout).Seconds,
		"route.timeout.seconds",
		"",
		"timeout time in seconds")
	flags.StringVar(&(rrOpts.Timeout).Nanos,
		"route.timeout.nanos",
		"",
		"timeout time in nanoseconds")
}

func AddRetryFlags(cmd *cobra.Command, opts *options.Options) {
	rrOpts := &(opts.Create).InputRoutingRule
	flags := cmd.Flags()
	flags.StringVar(&(rrOpts.Retry).Attempts,
		"route.retry.attempt",
		"",
		"number of times to retry")
	flags.StringVar(&(rrOpts.Retry.PerTryTimeout).Seconds,
		"route.retry.timeout.seconds",
		"",
		"retry timeout time in seconds")
	flags.StringVar(&(rrOpts.Retry.PerTryTimeout).Nanos,
		"route.retry.timeout.nanos",
		"",
		"retry timeout time in nanoseconds")
}

func AddFaultFlags(cmd *cobra.Command, opts *options.Options) {
	rrOpts := &(opts.Create).InputRoutingRule
	flags := cmd.Flags()
	// delay
	flags.StringVar(&(rrOpts.FaultInjection).DelayPercent,
		"fault.delay.percent",
		"",
		"Percentage of requests on which the delay will be injected (0-100).")
	flags.StringVar(&(rrOpts.FaultInjection).HttpDelayType,
		"fault.delay.type",
		"",
		"Type of delay (fixed or exponential).")
	flags.StringVar(&(rrOpts.FaultInjection).HttpDelayValue.Seconds,
		"fault.delay.value.seconds",
		"",
		"delay duration (seconds).")
	flags.StringVar(&(rrOpts.FaultInjection).HttpDelayValue.Nanos,
		"fault.delay.value.nanos",
		"",
		"delay duration (nanoseconds).")
	// abort
	flags.StringVar(&(rrOpts.FaultInjection).AbortPercent,
		"fault.abort.percent",
		"",
		"Percentage of requests on which the abort will be injected (0-100).")
	flags.StringVar(&(rrOpts.FaultInjection).ErrorType,
		"fault.abort.type",
		"",
		"Type of error (http, http2, or grpc).")
	flags.StringVar(&(rrOpts.FaultInjection).ErrorMessage,
		"fault.abort.message",
		"",
		"Error message (int for type=http errors, string otherwise).")
}

func AddTrafficShiftingFlags(cmd *cobra.Command, opts *options.Options) {
	tsOpts := &(opts.Create.InputRoutingRule).TrafficShifting
	flags := cmd.Flags()

	flags.StringVar(&tsOpts.Upstreams,
		"traffic.upstreams",
		"",
		"Upstreams for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon.")

	flags.StringVar(&tsOpts.Weights,
		"traffic.weights",
		"",
		"Comma-separated list of integer weights corresponding to the associated upstream's traffic sharing percentage.")

}

func AddCorsFlags(cmd *cobra.Command, opts *options.Options) {
	cOpts := &(opts.Create.InputRoutingRule).Cors
	flags := cmd.Flags()

	flags.StringVar(&cOpts.AllowOrigin,
		"cors.allow.origin",
		"",
		"The list of origins that are allowed to perform CORS requests. The content will be serialized into the Access-Control-Allow-Origin header. Wildcard * will allow all origins.")

	flags.StringVar(&cOpts.AllowMethods,
		"cors.allow.methods",
		"",
		"List of HTTP methods allowed to access the resource. The content will be serialized into the Access-Control-Allow-Methods header.")

	flags.StringVar(&cOpts.AllowHeaders,
		"cors.allow.headers",
		"",
		"List of HTTP headers that can be used when requesting the resource. Serialized to Access-Control-Allow-Methods header.")

	flags.StringVar(&cOpts.ExposeHeaders,
		"cors.expose.headers",
		"",
		"A white list of HTTP headers that the browsers are allowed to access. Serialized into Access-Control-Expose-Headers header.")

	flags.StringVar(&(cOpts.MaxAge).Seconds,
		"cors.maxage.seconds",
		"",
		"Max age time in seconds. Specifies how long the the results of a preflight request can be cached. Translates to the Access-Control-Max-Age header.")

	flags.StringVar(&(cOpts.MaxAge).Nanos,
		"cors.maxage.nanos",
		"",
		"Max age time in nanoseconds. Specifies how long the the results of a preflight request can be cached. Translates to the Access-Control-Max-Age header.")

	flags.BoolVar(&cOpts.AllowCredentials,
		"cors.allow.credentials",
		false,
		"Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials. Translates to Access-Control-Allow-Credentials header.")

}

func AddMirrorFlags(cmd *cobra.Command, opts *options.Options) {
	mOpts := &(opts.Create.InputRoutingRule).Mirror
	flags := cmd.Flags()

	flags.StringVar(&mOpts.Upstream,
		"mirror",
		"",
		"Destination upstream (ex: upstream_namespace:upstream_name).")
}

func AddHeaderManipulationFlags(cmd *cobra.Command, opts *options.Options) {
	hOpts := &(opts.Create.InputRoutingRule).HeaderManipulation
	flags := cmd.Flags()

	flags.StringVar(&hOpts.RemoveResponseHeaders,
		"header.response.remove",
		"",
		"Headers to remove from response (ex: h1,h2).")
	flags.StringVar(&hOpts.AppendResponseHeaders,
		"header.response.append",
		"",
		"Headers to append to response (ex: h1,v1,h2,v2).")
	flags.StringVar(&hOpts.RemoveRequestHeaders,
		"header.request.remove",
		"",
		"Headers to remove from request (ex: h1,h2).")
	flags.StringVar(&hOpts.AppendRequestHeaders,
		"header.request.append",
		"",
		"Headers to append to request (ex: h1,v1,h2,v2).")
}
