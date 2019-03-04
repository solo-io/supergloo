package options

import (
	"context"

	"github.com/solo-io/gloo/projects/gloo/cli/pkg/cmd/options"

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
	InputInstall InputInstall
}

type InputInstall struct {
	IstioInstall v1.Install_Istio
}

type Uninstall struct {
	Metadata core.Metadata
}

type CreateRoutingRule struct {
	SourceSelector      Selector
	DestinationSelector Selector
	RequestMatchers     []options.RouteMatchers
	RoutingRuleSpec     RoutingRuleSpec
}

const (
	SelectorType_Labels    = "Label Selector"
	SelectorType_Namespace = "Namespace Selector"
	SelectorType_Upstream  = "Upstream Selector"
)

type Selector struct {
	Enabled            bool // shows that this selecotr is non-zero
	SelectorType       string
	SelectedUpstreams  []core.ResourceRef
	SelectedNamespaces []string
	SelectedLabels     map[string]string
}

// no implemented specs yet
type RoutingRuleSpec struct {
	SpecType string
}
