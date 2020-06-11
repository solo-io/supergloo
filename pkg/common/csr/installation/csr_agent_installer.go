package installation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/version"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/helm"
)

var (
	FailedToSetUpCsrAgent = func(err error) error {
		return eris.Wrap(err, "Could not set up CSR agent on remote cluster")
	}
)

const (
	CsrAgentChartUriTemplate     = "https://storage.googleapis.com/service-mesh-hub/csr-agent/csr-agent-%s.tgz"
	LocallyPackagedChartTemplate = "./_output/helm/charts/csr-agent/csr-agent-%s.tgz"
)

func NewCsrAgentInstallerFactory(
	helmClientFileConfigFactory helm.HelmClientForFileConfigFactory,
	helmClientMemoryConfigFactory helm.HelmClientForMemoryConfigFactory,
	deployedVersionFinder version.DeployedVersionFinder,
) CsrAgentInstallerFactory {
	return func(
		helmInstallerFactory helm.HelmInstallerFactory,
	) CsrAgentInstaller {
		return NewCsrAgentInstaller(
			helmClientFileConfigFactory,
			helmClientMemoryConfigFactory,
			deployedVersionFinder,
			helmInstallerFactory,
		)
	}
}

func NewCsrAgentInstaller(
	helmClientFileConfigFactory helm.HelmClientForFileConfigFactory,
	helmClientMemoryConfigFactory helm.HelmClientForMemoryConfigFactory,
	deployedVersionFinder version.DeployedVersionFinder,
	helmInstallerFactory helm.HelmInstallerFactory,
) CsrAgentInstaller {
	return &csrAgentInstaller{
		helmClientFileConfigFactory:   helmClientFileConfigFactory,
		helmClientMemoryConfigFactory: helmClientMemoryConfigFactory,
		helmInstallerFactory:          helmInstallerFactory,
		deployedVersionFinder:         deployedVersionFinder,
	}
}

type csrAgentInstaller struct {
	helmClientFileConfigFactory   helm.HelmClientForFileConfigFactory
	helmClientMemoryConfigFactory helm.HelmClientForMemoryConfigFactory
	helmInstallerFactory          helm.HelmInstallerFactory
	deployedVersionFinder         version.DeployedVersionFinder
}

func (c *csrAgentInstaller) Install(
	ctx context.Context,
	installOptions *CsrAgentInstallOptions,
) error {
	openSourceVersion, err := c.deployedVersionFinder.OpenSourceVersion(ctx, installOptions.SmhInstallNamespace)
	if err != nil {
		return FailedToSetUpCsrAgent(err)
	}

	err = c.runHelmInstall(openSourceVersion, installOptions)

	// if we already have a CSR agent running here, then we're done
	if eris.Is(err, helminstall.ReleaseAlreadyInstalledErr(installOptions.ReleaseName, installOptions.RemoteWriteNamespace)) {
		return nil
	} else if err != nil {
		return FailedToSetUpCsrAgent(err)
	}

	return nil
}

func (c *csrAgentInstaller) Uninstall(uninstallOptions *CsrAgentUninstallOptions) error {
	helmClient := c.buildHelmClient(uninstallOptions.KubeConfig)
	uninstaller, err := helmClient.NewUninstall(uninstallOptions.ReleaseNamespace)
	if err != nil {
		return err
	}
	_, err = uninstaller.Run(uninstallOptions.ReleaseName)
	return err
}

func (c *csrAgentInstaller) runHelmInstall(
	version string,
	opts *CsrAgentInstallOptions,
) error {
	var chartPathTemplate string
	if opts.UseDevCsrAgentChart {
		chartPathTemplate = LocallyPackagedChartTemplate
	} else {
		chartPathTemplate = CsrAgentChartUriTemplate
	}
	releaseUri := fmt.Sprintf(chartPathTemplate, version)
	helmClient := c.buildHelmClient(opts.KubeConfig)
	return c.helmInstallerFactory(helmClient).Install(&types.InstallerConfig{
		InstallNamespace: opts.RemoteWriteNamespace,
		CreateNamespace:  true,
		ReleaseName:      opts.ReleaseName,
		ReleaseUri:       releaseUri,
		ExtraValues:      opts.ExtraValues,
	})
}

func (c *csrAgentInstaller) buildHelmClient(kubeConfig KubeConfig) types.HelmClient {
	if kubeConfig.KubeConfig != nil {
		return c.helmClientMemoryConfigFactory(kubeConfig.KubeConfig)
	} else {
		return c.helmClientFileConfigFactory(kubeConfig.KubeConfigPath, kubeConfig.KubeContext)
	}
}
