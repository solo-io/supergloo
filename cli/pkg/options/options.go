package options

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type Options struct {
	// common
	Ctx         context.Context
	Interactive bool
	OutputType  string
	Metadata    core.Metadata

	Init              Init
	Install           Install
	Uninstall         Uninstall
	CreateRoutingRule CreateRoutingRule
}

type Init struct {
	HelmChartOverride string
	HelmValues        string
	InstallNamespace  string
	ReleaseVersion    string
	DryRun            bool
}

type Install struct {
	Update       bool // if install exists and is enabled, update with new opts
	IstioInstall v1.Install_Istio
}

type Uninstall struct {
	Metadata core.Metadata
}

type CreateRoutingRule struct {
	SourceSelector      Selector
	DestinationSelector Selector
	RequestMatchers     RequestMatchersValue
	RoutingRuleSpec     RoutingRuleSpec
}

type RequestMatcher struct {
	PathPrefix    string            `json:"path_prefix"`
	PathExact     string            `json:"path_exact"`
	PathRegex     string            `json:"path_regex"`
	Methods       []string          `json:"methods"`
	HeaderMatcher map[string]string `json:"header_matchers"`
}

type Selector struct {
	SelectedUpstreams  ResourceRefsValue
	SelectedNamespaces []string
	SelectedLabels     MapStringStringValue
}

// no implemented specs yet
type RoutingRuleSpec struct {
	TrafficShifting TrafficShiftingValue
}
