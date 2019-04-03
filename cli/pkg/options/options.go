package options

import (
	"context"
	"time"

	glooOptions "github.com/solo-io/gloo/projects/gloo/cli/pkg/cmd/options"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type Options struct {
	// common
	Ctx           context.Context
	Interactive   bool
	OutputType    string
	Metadata      core.Metadata
	PrintKubeYaml bool

	Init               Init
	Install            Install
	Uninstall          Uninstall
	CreateRoutingRule  CreateRoutingRule
	CreateSecurityRule CreateSecurityRule
	CreateTlsSecret    CreateTlsSecret
	CreateAwsSecret    CreateAwsSecret
	EditUpstream       EditUpstream
	GetMeshIngress     GetMeshIngress
	SetRootCert        SetRootCert
	SetStats           SetStats
	RegisterAppMesh    RegisterAppMesh
}

type Init struct {
	HelmChartOverride string
	HelmValues        string
	InstallNamespace  string
	ReleaseVersion    string
	DryRun            bool
}

type InstallationNamespace struct {
	Istio string
	Gloo  string
}

type Install struct {
	Update                bool // if install exists and is enabled, update with new opts
	InstallationNamespace InstallationNamespace
	IstioInstall          v1.IstioInstall
	GlooIngressInstall    v1.GlooInstall
	MeshIngress           MeshIngressInstall
}

type MeshIngressInstall struct {
	Meshes ResourceRefsValue
}

type Uninstall struct {
	Metadata core.Metadata
}

type CreateRoutingRule struct {
	SourceSelector      Selector
	DestinationSelector Selector
	TargetMesh          ResourceRefValue
	RequestMatchers     RequestMatchersValue
	RoutingRuleSpec     RoutingRuleSpec
}

type CreateSecurityRule struct {
	SourceSelector      Selector
	DestinationSelector Selector
	TargetMesh          ResourceRefValue
	AllowedMethods      []string
	AllowedPaths        []string
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
	FaultInjection  FaultInjection
}

type FaultInjection struct {
	Percent float64
	Abort   FaultInjectionAbort
	Delay   FaultInjectionDelay
}
type FaultInjectionDelay struct {
	Fixed time.Duration
}
type FaultInjectionAbort struct {
	Http v1.FaultInjection_Abort_HttpStatus
}

type CreateTlsSecret struct {
	RootCaFilename     string
	PrivateKeyFilename string
	CertChainFilename  string
	CaCertFilename     string
}

type CreateAwsSecret struct {
	CredentialsFileLocation string
	CredentialsFileProfile  string
	AccessKeyId             string
	SecretAccessKey         string
}

type SetRootCert struct {
	TargetMesh ResourceRefValue
	TlsSecret  ResourceRefValue
}

type EditUpstream struct {
	MtlsMesh ResourceRefValue
}

type SetStats struct {
	TargetMesh           ResourceRefValue
	PrometheusConfigMaps ResourceRefsValue
}

type GetMeshIngress struct {
	Proxy  glooOptions.Proxy
	Target ResourceRefValue
}

type RegisterAppMesh struct {
	Region              string
	Secret              ResourceRefValue
	EnableAutoInjection string
	ConfigMap           ResourceRefValue
	PodSelector         Selector
	VirtualNodeLabel    string
}
