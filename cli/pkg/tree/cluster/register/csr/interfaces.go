package csr

import (
	"context"

	"github.com/solo-io/go-utils/installutils/helminstall"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type CsrAgentInstallOptions struct {
	// File KubeConfig for remote cluster
	KubeConfigPath string
	KubeContext    string
	// In memory KubeConfig for remote cluster
	KubeConfig clientcmd.ClientConfig

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

type CsrAgentInstallerFactory func(helmInstallerFactory helminstall.InstallerFactory) CsrAgentInstaller
