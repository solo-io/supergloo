package csr

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type CsrAgentInstallOptions struct {
	// Either KubeConfig or KubeconfigPath + KubeContext needs to be provided. KubeConfig takes precedence if both are supplied.
	// In memory KubeConfig for remote cluster.
	KubeConfig clientcmd.ClientConfig
	// File KubeConfig for remote cluster.
	KubeConfigPath string
	KubeContext    string

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

type CsrAgentInstallerFactory func(helmInstallerFactory factories.HelmerInstallerFactory) CsrAgentInstaller
