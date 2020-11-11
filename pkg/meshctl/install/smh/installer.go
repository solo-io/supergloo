package smh

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/helm"
)

const (
	ServiceMeshHubChartUriTemplate = "https://storage.googleapis.com/service-mesh-hub/service-mesh-hub/service-mesh-hub-%s.tgz"
	CertAgentChartUriTemplate      = "https://storage.googleapis.com/service-mesh-hub/cert-agent/cert-agent-%s.tgz"
	WasmAgentChartUriTemplate      = "https://storage.googleapis.com/service-mesh-hub/wasm-agent/wasm-agent-%s.tgz"
	serviceMeshHubReleaseName      = "service-mesh-hub"
	certAgentReleaseName           = "cert-agent"
	wasmAgentReleaseName           = "wasm-agent"
)

type Installer struct {
	HelmChartPath  string
	HelmValuesPath string
	KubeConfig     string
	KubeContext    string
	Namespace      string
	ReleaseName    string
	Verbose        bool
	DryRun         bool
}

func (i Installer) InstallServiceMeshHub(
	ctx context.Context,
) error {
	return i.install(ctx, ServiceMeshHubChartUriTemplate, serviceMeshHubReleaseName)
}

func (i Installer) InstallCertAgent(
	ctx context.Context,
) error {
	return i.install(ctx, CertAgentChartUriTemplate, certAgentReleaseName)
}

func (i Installer) InstallWasmAgent(
	ctx context.Context,
) error {
	return i.install(ctx, WasmAgentChartUriTemplate, wasmAgentReleaseName)
}

func (i Installer) install(
	ctx context.Context,
	chartUriTemplate string,
	releaseName string,
) error {
	helmChartOverride := i.HelmChartPath

	if i.ReleaseName != "" {
		releaseName = i.ReleaseName
	}

	if helmChartOverride == "" {
		helmChartOverride = fmt.Sprintf(chartUriTemplate, version.Version)
	}

	return helm.Installer{
		KubeConfig:  i.KubeConfig,
		KubeContext: i.KubeContext,
		ChartUri:    helmChartOverride,
		Namespace:   i.Namespace,
		ReleaseName: releaseName,
		ValuesFile:  i.HelmValuesPath,
		Verbose:     i.Verbose,
		DryRun:      i.DryRun,
	}.InstallChart(ctx)
}
