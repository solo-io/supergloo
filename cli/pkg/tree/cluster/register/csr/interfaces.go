package csr

import (
	"context"

	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/solo-io/service-mesh-hub/pkg/version"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type CsrAgentInstallOptions struct {
	// these two kube fields should be for the *remote* cluster
	KubeConfig  string
	KubeContext string

	ClusterName          string
	SmhInstallNamespace  string
	UseDevCsrAgentChart  bool // if true, look for a chart in .tgz form in _output/helm/charts/csr-agent
	ReleaseName          string
	RemoteWriteNamespace string
}

type CsrAgentInstaller interface {
	Install(
		ctx context.Context,
		installOptions *CsrAgentInstallOptions,
	) error
}

type CsrAgentInstallerFactory func(
	helmInstaller types.Installer,
	deployedVersionFinder version.DeployedVersionFinder,
) CsrAgentInstaller
