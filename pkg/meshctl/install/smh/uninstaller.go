package smh

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/helm"
)

type Uninstaller struct {
	KubeConfig  string
	KubeContext string
	Namespace   string
	ReleaseName string
	Verbose     bool
	DryRun      bool
}

func (i Uninstaller) UninstallServiceMeshHub(
	ctx context.Context,
) error {
	return i.uninstall(ctx, serviceMeshHubReleaseName)
}

func (i Uninstaller) UninstallCertAgent(
	ctx context.Context,
) error {
	return i.uninstall(ctx, certAgentReleaseName)
}

func (i Uninstaller) uninstall(
	ctx context.Context,
	releaseName string,
) error {
	if i.ReleaseName != "" {
		releaseName = i.ReleaseName
	}

	return helm.Uninstaller{
		KubeConfig:  i.KubeConfig,
		KubeContext: i.KubeContext,
		Namespace:   i.Namespace,
		ReleaseName: releaseName,
		Verbose:     i.Verbose,
		DryRun:      i.DryRun,
	}.UninstallChart(ctx)
}
