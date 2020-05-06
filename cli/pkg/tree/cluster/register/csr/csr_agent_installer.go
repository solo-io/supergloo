package csr

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"github.com/solo-io/service-mesh-hub/pkg/version"
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
	helmClientFileConfigFactory factories.HelmClientForFileConfigFactory,
	helmClientMemoryConfigFactory factories.HelmClientForMemoryConfigFactory,
	deployedVersionFinder version.DeployedVersionFinder,
) CsrAgentInstallerFactory {
	return func(
		helmInstallerFactory helminstall.InstallerFactory,
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
	helmClientFileConfigFactory factories.HelmClientForFileConfigFactory,
	helmClientMemoryConfigFactory factories.HelmClientForMemoryConfigFactory,
	deployedVersionFinder version.DeployedVersionFinder,
	helmInstallerFactory helminstall.InstallerFactory,
) CsrAgentInstaller {
	return &csrAgentInstaller{
		helmClientFileConfigFactory:   helmClientFileConfigFactory,
		helmClientMemoryConfigFactory: helmClientMemoryConfigFactory,
		helmInstallerFactory:          helmInstallerFactory,
		deployedVersionFinder:         deployedVersionFinder,
	}
}

type csrAgentInstaller struct {
	helmClientFileConfigFactory   factories.HelmClientForFileConfigFactory
	helmClientMemoryConfigFactory factories.HelmClientForMemoryConfigFactory
	helmInstallerFactory          helminstall.InstallerFactory
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

func (c *csrAgentInstaller) runHelmInstall(
	version string,
	opts *CsrAgentInstallOptions,
) error {
	var chartPathTemplate string
	var helmClient types.HelmClient
	if opts.UseDevCsrAgentChart {
		chartPathTemplate = LocallyPackagedChartTemplate
	} else {
		chartPathTemplate = CsrAgentChartUriTemplate
	}
	releaseUri := fmt.Sprintf(chartPathTemplate, version)
	if opts.KubeConfig != nil {
		helmClient = c.helmClientMemoryConfigFactory(opts.KubeConfig)
	} else {
		helmClient = c.helmClientFileConfigFactory(opts.KubeConfigPath, opts.KubeContext)
	}
	return c.helmInstallerFactory(helmClient).Install(&types.InstallerConfig{
		InstallNamespace: opts.RemoteWriteNamespace,
		CreateNamespace:  true,
		ReleaseName:      opts.ReleaseName,
		ReleaseUri:       releaseUri,
	})
}
