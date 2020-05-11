package cluster_registration

import (
	"context"
	"io/ioutil"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/config_lookup"
	crd_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/crd"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_core_v1_clients "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	zephyr_security_scheme "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"github.com/solo-io/service-mesh-hub/pkg/installers/csr"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/security/secrets"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	FailedToCleanUpKubeConfigSecret = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to clean up kubeconfig secret for cluster %s", clusterName)
	}
	FailedToCleanUpKubeConfigCrd = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to clean up kubeconfig CRD for cluster %s", clusterName)
	}
	FailedToCleanUpCertSecrets = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to clean up TLS cert data for cluster %s", clusterName)
	}
	FailedToCleanUpServiceAccount = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to clean up Service Mesh Hub service account from cluster %s", clusterName)
	}

	noOpHelmLogger = func(format string, v ...interface{}) {}
)

func NewClusterDeregistrationClient(
	crdRemover crd_uninstall.CrdRemover,
	csrAgentInstallerFactory csr.CsrAgentInstallerFactory,
	kubeConfigLookup config_lookup.KubeConfigLookup,
	localkubeClusterClient zephyr_discovery.KubernetesClusterClient,
	localSecretClient k8s_core_v1_clients.SecretClient,
	secretClientFactory k8s_core_v1_clients.SecretClientFactory,
	dynamicClientGetter mc_manager.DynamicClientGetter,
	serviceAccountClientFactory k8s_core_v1_clients.ServiceAccountClientFactory,
) ClusterDeregistrationClient {
	return &clusterDeregistrationClient{
		crdRemover:                  crdRemover,
		csrAgentInstallerFactory:    csrAgentInstallerFactory,
		kubeConfigLookup:            kubeConfigLookup,
		localkubeClusterClient:      localkubeClusterClient,
		localSecretClient:           localSecretClient,
		secretClientFactory:         secretClientFactory,
		dynamicClientGetter:         dynamicClientGetter,
		serviceAccountClientFactory: serviceAccountClientFactory,
	}
}

type clusterDeregistrationClient struct {
	crdRemover                  crd_uninstall.CrdRemover
	kubeLoader                  kubeconfig.KubeLoader
	csrAgentInstallerFactory    csr.CsrAgentInstallerFactory
	kubeConfigLookup            config_lookup.KubeConfigLookup
	localkubeClusterClient      zephyr_discovery.KubernetesClusterClient
	localSecretClient           k8s_core_v1_clients.SecretClient
	secretClientFactory         k8s_core_v1_clients.SecretClientFactory
	serviceAccountClientFactory k8s_core_v1_clients.ServiceAccountClientFactory
	dynamicClientGetter         mc_manager.DynamicClientGetter
}

func (c *clusterDeregistrationClient) Deregister(ctx context.Context, kubeCluster *zephyr_discovery.KubernetesCluster) error {
	config, err := c.kubeConfigLookup.FromCluster(ctx, kubeCluster.GetName())
	if meta.IsNoMatchError(err) {
		return nil
	} else if err != nil {
		return FailedToFindClusterCredentials(err, kubeCluster.GetName())
	}
	kubeClient := kubernetes.NewForConfigOrDie(config.RestConfig)
	helmInstallerFactory := factories.NewHelmInstallerFactory(kubeClient.CoreV1().Namespaces(), ioutil.Discard)
	csrAgentInstaller := c.csrAgentInstallerFactory(helmInstallerFactory)
	err = csrAgentInstaller.Uninstall(&csr.CsrAgentUninstallOptions{
		KubeConfig:       csr.KubeConfig{KubeConfig: config.ClientConfig},
		ReleaseName:      cliconstants.CsrAgentReleaseName,
		ReleaseNamespace: kubeCluster.Spec.GetWriteNamespace(),
	})
	if err != nil {
		return FailedToUninstallCsrAgent(err, kubeCluster.GetName())
	}

	clientForCluster, err := c.dynamicClientGetter.GetClientForCluster(ctx, kubeCluster.GetName())
	if err != nil {
		return err
	}

	err = c.cleanUpCertSecrets(ctx, clientForCluster, kubeCluster)
	if err != nil {
		return FailedToCleanUpCertSecrets(err, kubeCluster.GetName())
	}

	err = c.cleanUpKubeConfigSecret(ctx, kubeCluster)
	if err != nil {
		return FailedToCleanUpKubeConfigSecret(err, kubeCluster.GetName())
	}

	err = c.localkubeClusterClient.DeleteKubernetesCluster(ctx, clients.ObjectMetaToObjectKey(kubeCluster.ObjectMeta))
	if err != nil {
		return FailedToCleanUpKubeConfigCrd(err, kubeCluster.GetName())
	}

	err = c.cleanUpServiceAccounts(ctx, clientForCluster, kubeCluster)
	if err != nil {
		return FailedToCleanUpServiceAccount(err, kubeCluster.GetName())
	}

	// the CSR agent installs only CRDs from the security group. Remove only those
	_, err = c.crdRemover.RemoveCrdGroup(ctx, kubeCluster.GetName(), config.RestConfig, zephyr_security_scheme.SchemeGroupVersion)
	if err != nil && !meta.IsNoMatchError(err) {
		return FailedToRemoveCrds(err, kubeCluster.GetName())
	}

	return nil
}

func (c *clusterDeregistrationClient) cleanUpServiceAccounts(ctx context.Context, clientForCluster client.Client, kubeCluster *zephyr_discovery.KubernetesCluster) error {
	serviceAccountClient := c.serviceAccountClientFactory(clientForCluster)
	err := serviceAccountClient.DeleteAllOfServiceAccount(
		ctx,
		client.InNamespace(kubeCluster.Spec.GetWriteNamespace()),
		client.MatchingLabels{
			cliconstants.ManagedByLabel:     cliconstants.ServiceMeshHubApplicationName,
			auth.RegistrationServiceAccount: auth.RegistrationServiceAccountValue,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *clusterDeregistrationClient) cleanUpCertSecrets(ctx context.Context, clientForCluster client.Client, kubeCluster *zephyr_discovery.KubernetesCluster) error {
	secretClientForCluster := c.secretClientFactory(clientForCluster)
	allSecrets, err := secretClientForCluster.ListSecret(ctx, client.InNamespace(kubeCluster.Spec.GetWriteNamespace()))
	if err != nil {
		return err
	}

	for _, secretIter := range allSecrets.Items {
		secret := secretIter

		if secret.Type == cert_secrets.IntermediateCertSecretType {
			err := secretClientForCluster.DeleteSecret(ctx, clients.ObjectMetaToObjectKey(secret.ObjectMeta))

			// if we have multiple de-registrations going on at once (potentially in `meshctl uninstall`, ignore the error if something else cleaned up the secret first)
			if err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (c *clusterDeregistrationClient) cleanUpKubeConfigSecret(ctx context.Context, kubeCluster *zephyr_discovery.KubernetesCluster) error {
	kubeConfigSecret, err := c.localSecretClient.GetSecret(ctx, clients.ResourceRefToObjectKey(kubeCluster.Spec.GetSecretRef()))
	if err != nil {
		return err
	}

	err = c.localSecretClient.DeleteSecret(ctx, clients.ObjectMetaToObjectKey(kubeConfigSecret.ObjectMeta))
	if err != nil {
		return err
	}

	return nil
}
