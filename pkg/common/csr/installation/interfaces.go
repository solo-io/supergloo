package installation

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/common/kube/helm"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type CsrAgentInstallOptions struct {
	KubeConfig           KubeConfig
	SmhInstallNamespace  string
	UseDevCsrAgentChart  bool // if true, look for a chart in .tgz form in _output/helm/charts/csr-agent
	ReleaseName          string
	RemoteWriteNamespace string
	ExtraValues          map[string]interface{}
}

type CsrAgentUninstallOptions struct {
	KubeConfig       KubeConfig
	ReleaseName      string
	ReleaseNamespace string
}

type KubeConfig struct {
	// Either KubeConfig or KubeconfigPath + KubeContext needs to be provided. KubeConfig takes precedence if both are supplied.
	KubeConfig     clientcmd.ClientConfig // in memory kube config
	KubeConfigPath string                 // on disk kube config
	KubeContext    string
}

type CsrAgentInstaller interface {
	Install(ctx context.Context, installOptions *CsrAgentInstallOptions) error
	Uninstall(uninstallOptions *CsrAgentUninstallOptions) error
}

type CsrAgentInstallerFactory func(helmInstallerFactory helm.HelmInstallerFactory) CsrAgentInstaller
