package smh

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/helm"
	"k8s.io/client-go/tools/clientcmd"
)

type Uninstaller struct {
	KubeConfig  clientcmd.ClientConfig
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
		Namespace:   i.Namespace,
		ReleaseName: releaseName,
		Verbose:     i.Verbose,
		DryRun:      i.DryRun,
	}.UninstallChart(ctx)
}
