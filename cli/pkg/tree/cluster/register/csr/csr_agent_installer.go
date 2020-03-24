package csr

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/solo-io/mesh-projects/pkg/version"
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

func NewCsrAgentInstallerFactory() CsrAgentInstallerFactory {
	return NewCsrAgentInstaller
}

func NewCsrAgentInstaller(
	helmInstaller types.Installer,
	deployedVersionFinder version.DeployedVersionFinder,
) CsrAgentInstaller {
	return &csrAgentInstaller{
		helmInstaller:         helmInstaller,
		deployedVersionFinder: deployedVersionFinder,
	}
}

type csrAgentInstaller struct {
	helmInstaller         types.Installer
	deployedVersionFinder version.DeployedVersionFinder
}

func (c *csrAgentInstaller) Install(
	ctx context.Context,
	installOptions *CsrAgentInstallOptions,
) error {
	openSourceVersion, err := c.deployedVersionFinder.OpenSourceVersion(ctx, installOptions.SmhInstallNamespace)
	if err != nil {
		return FailedToSetUpCsrAgent(err)
	}

	err = c.runHelmInstall(
		installOptions.KubeConfig,
		installOptions.KubeContext,
		openSourceVersion,
		installOptions.RemoteWriteNamespace,
		installOptions.ReleaseName,
		installOptions.UseDevCsrAgentChart,
	)

	// if we already have a CSR agent running here, then we're done
	if eris.Is(err, helminstall.ReleaseAlreadyInstalledErr(installOptions.ReleaseName, installOptions.RemoteWriteNamespace)) {
		return nil
	} else if err != nil {
		return FailedToSetUpCsrAgent(err)
	}

	return nil
}

func (c *csrAgentInstaller) runHelmInstall(
	kubeConfig string,
	kubeContext string,
	version string,
	installNamespace string,
	releaseName string,
	useDevCsrAgentChart bool,
) error {

	var chartPathTemplate string
	if useDevCsrAgentChart {
		chartPathTemplate = LocallyPackagedChartTemplate
	} else {
		chartPathTemplate = CsrAgentChartUriTemplate
	}

	releaseUri := fmt.Sprintf(chartPathTemplate, version)

	return c.helmInstaller.Install(&types.InstallerConfig{
		KubeConfig:       kubeConfig,
		KubeContext:      kubeContext,
		InstallNamespace: installNamespace,
		CreateNamespace:  true,
		ReleaseName:      releaseName,
		ReleaseUri:       releaseUri,
	})
}
