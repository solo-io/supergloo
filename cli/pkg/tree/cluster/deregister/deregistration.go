package deregister

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	crd_uninstall "github.com/solo-io/mesh-projects/cli/pkg/tree/uninstall/crd"
	helm_uninstall "github.com/solo-io/mesh-projects/cli/pkg/tree/uninstall/helm"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	FailedToFindKubeConfigSecret = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to find kube config secret for cluster %s", clusterName)
	}
	FailedToConvertSecretToKubeConfig = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to convert kube config secret for cluster %s to REST config", clusterName)
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
	secretsClient kubernetes_core.SecretsClient,
	secretToConfigConverter kubeconfig.SecretToConfigConverter,
	crdRemover crd_uninstall.CrdRemover,
	inMemoryRESTClientFactory common_config.InMemoryRESTClientGetterFactory,
	helmUninstallerClientFactory helm_uninstall.UninstallerFactory,
) ClusterDeregistrationClient {
	return &clusterDeregistrationClient{
		secretsClient:                secretsClient,
		secretToConfigConverter:      secretToConfigConverter,
		crdRemover:                   crdRemover,
		inMemoryRESTClientFactory:    inMemoryRESTClientFactory,
		helmUninstallerClientFactory: helmUninstallerClientFactory,
	}
}

type clusterDeregistrationClient struct {
	secretsClient                kubernetes_core.SecretsClient
	secretToConfigConverter      kubeconfig.SecretToConfigConverter
	crdRemover                   crd_uninstall.CrdRemover
	kubeLoader                   common_config.KubeLoader
	inMemoryRESTClientFactory    common_config.InMemoryRESTClientGetterFactory
	helmUninstallerClientFactory helm_uninstall.UninstallerFactory
}

func (c *clusterDeregistrationClient) Run(ctx context.Context, kubeCluster *discovery_v1alpha1.KubernetesCluster) error {
	cfgSecretRef := kubeCluster.Spec.GetSecretRef()
	secret, err := c.secretsClient.Get(ctx, cfgSecretRef.GetName(), cfgSecretRef.GetNamespace())
	if err != nil {
		return FailedToFindKubeConfigSecret(err, kubeCluster.GetName())
	}

	clusterName, config, err := c.secretToConfigConverter(secret)
	if err != nil {
		return FailedToConvertSecretToKubeConfig(err, kubeCluster.GetName())
	}

	restClientGetter := c.inMemoryRESTClientFactory(config.RestConfig)
	helmRestClientGetter := &helmRESTClientGetter{
		baseRESTClientGetter: restClientGetter,
		clientConfig:         config.ClientConfig,
	}

	helmUninstaller, err := c.helmUninstallerClientFactory(helmRestClientGetter, kubeCluster.Spec.GetWriteNamespace(), noOpHelmLogger)
	if err != nil {
		return FailedToSetUpHelmUnintaller(err, clusterName)
	}

	_, err = helmUninstaller.Run(cliconstants.CsrAgentReleaseName)
	if err != nil {
		return FailedToUninstallCsrAgent(err, kubeCluster.GetName())
	}

	_, err = c.crdRemover.RemoveZephyrCrds(clusterName, config.RestConfig)
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
