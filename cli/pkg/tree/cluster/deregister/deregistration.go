package deregister

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/config_lookup"
	crd_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/crd"
	helm_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/helm"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	FailedToFindClusterCredentials = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to load credentials for cluster %s", clusterName)
	}
	FailedToUninstallCsrAgent = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to uninstall CSR agent on cluster %s", clusterName)
	}
	FailedToRemoveCrds = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to remove CRDs from cluster %s", clusterName)
	}
	FailedToSetUpHelmUnintaller = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to set up Helm uninstaller %s", clusterName)
	}

	noOpHelmLogger = func(format string, v ...interface{}) {}
)

func NewClusterDeregistrationClient(
	crdRemover crd_uninstall.CrdRemover,
	inMemoryRESTClientFactory common_config.InMemoryRESTClientGetterFactory,
	helmUninstallerClientFactory helm_uninstall.UninstallerFactory,
	kubeConfigLookup config_lookup.KubeConfigLookup,
) ClusterDeregistrationClient {
	return &clusterDeregistrationClient{
		crdRemover:                   crdRemover,
		inMemoryRESTClientFactory:    inMemoryRESTClientFactory,
		helmUninstallerClientFactory: helmUninstallerClientFactory,
		kubeConfigLookup:             kubeConfigLookup,
	}
}

type clusterDeregistrationClient struct {
	crdRemover                   crd_uninstall.CrdRemover
	kubeLoader                   common_config.KubeLoader
	inMemoryRESTClientFactory    common_config.InMemoryRESTClientGetterFactory
	helmUninstallerClientFactory helm_uninstall.UninstallerFactory
	kubeConfigLookup             config_lookup.KubeConfigLookup
}

func (c *clusterDeregistrationClient) Run(ctx context.Context, kubeCluster *discovery_v1alpha1.KubernetesCluster) error {
	config, err := c.kubeConfigLookup.FromCluster(ctx, kubeCluster.GetName())
	if err != nil {
		return FailedToFindClusterCredentials(err, kubeCluster.GetName())
	}

	restClientGetter := c.inMemoryRESTClientFactory(config.RestConfig)
	helmRestClientGetter := &helmRESTClientGetter{
		baseRESTClientGetter: restClientGetter,
		clientConfig:         config.ClientConfig,
	}

	helmUninstaller, err := c.helmUninstallerClientFactory(helmRestClientGetter, kubeCluster.Spec.GetWriteNamespace(), noOpHelmLogger)
	if err != nil {
		return FailedToSetUpHelmUnintaller(err, kubeCluster.GetName())
	}

	_, err = helmUninstaller.Run(cliconstants.CsrAgentReleaseName)
	if err != nil {
		return FailedToUninstallCsrAgent(err, kubeCluster.GetName())
	}

	_, err = c.crdRemover.RemoveZephyrCrds(kubeCluster.GetName(), config.RestConfig)
	if err != nil {
		return FailedToRemoveCrds(err, kubeCluster.GetName())
	}

	return nil
}

// Helm has their own RESTClientGetter, which has an extra method on top of `resource.RESTClientGetter`, because of course they do
// this type just delegates to the base RESTClientGetter and the extra method just returns the client config we already parsed out
type helmRESTClientGetter struct {
	baseRESTClientGetter resource.RESTClientGetter
	clientConfig         clientcmd.ClientConfig
}

func (h *helmRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return h.baseRESTClientGetter.ToRESTConfig()
}

func (h *helmRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return h.baseRESTClientGetter.ToDiscoveryClient()
}

func (h *helmRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return h.baseRESTClientGetter.ToRESTMapper()
}

func (h *helmRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return h.clientConfig
}
