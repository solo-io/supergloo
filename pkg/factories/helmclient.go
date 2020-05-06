package factories

import (
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"k8s.io/client-go/tools/clientcmd"
)

type HelmClientForFileConfigFactory func(kubeConfig, kubeContext string) types.HelmClient

type HelmClientForMemoryConfigFactory func(config clientcmd.ClientConfig) types.HelmClient

func HelmClientForFileConfigFactoryProvider() HelmClientForFileConfigFactory {
	return func(kubeConfig, kubeContext string) types.HelmClient {
		return helminstall.DefaultHelmClientFileConfig(kubeConfig, kubeContext)
	}
}

func HelmClientForMemoryConfigFactoryProvider() HelmClientForMemoryConfigFactory {
	return func(config clientcmd.ClientConfig) types.HelmClient {
		return helminstall.DefaultHelmClientMemoryConfig(config)
	}
}
