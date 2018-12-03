package options

import (
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

// InputRoutingRule is used to gather user input
type InputRoutingRule struct {
	RouteName          string
	TargetMesh         core.ResourceRef
	Sources            string
	Destinations       string
	Matchers           []string
	OverrideExisting   bool
	Timeout            InputDuration
	Retry              InputRetry
	FaultInjection     InputFaultInjection
	ActiveTypes        []MultiselectOptionBool
	TrafficShifting    InputTrafficShifting
	Cors               InputCors
	Mirror             InputMirror
	HeaderManipulation InputHeaderManipulation
}

type InputFaultInjection struct {
	DelayPercent string // int32

	// Options:
	//	*HTTPFaultInjection_Delay_FixedDelay // Duration
	//	*HTTPFaultInjection_Delay_ExponentialDelay // Duration
	HttpDelayType  string
	HttpDelayValue InputDuration

	AbortPercent string // int32

	// Options:
	//	*HTTPFaultInjection_Abort_HttpStatus // int32
	//	*HTTPFaultInjection_Abort_GrpcStatus // string
	//	*HTTPFaultInjection_Abort_Http2Error // string
	ErrorType    string
	ErrorMessage string
}

type MultiselectOptionBool struct {
	ID          string
	DisplayName string
	Active      bool
}

type InputTrafficShifting struct {
	Upstreams string
	Weights   string
}

type InputCors struct {
	AllowOrigin      string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	MaxAge           InputDuration
	AllowCredentials bool
}

type InputMirror struct {
	Upstream string
}

type InputHeaderManipulation struct {
	RemoveResponseHeaders string
	AppendResponseHeaders string
	RemoveRequestHeaders  string
	AppendRequestHeaders  string
}
