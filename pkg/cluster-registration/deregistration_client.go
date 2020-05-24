package cluster_registration

import (
	"context"
	"io/ioutil"

	"github.com/rotisserie/eris"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_core_v1_clients "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	zephyr_security_scheme "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/constants"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/csr/certgen/secrets"
	"github.com/solo-io/service-mesh-hub/pkg/csr/installation"
	"github.com/solo-io/service-mesh-hub/pkg/kube/auth"
	crd_uninstall "github.com/solo-io/service-mesh-hub/pkg/kube/crd"
	"github.com/solo-io/service-mesh-hub/pkg/kube/helm"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/kube/multicluster"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
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
)

func NewClusterDeregistrationClient(
	crdRemover crd_uninstall.CrdRemover,
	csrAgentInstallerFactory installation.CsrAgentInstallerFactory,
	kubeConfigLookup kubeconfig.KubeConfigLookup,
	localKubeClusterClient zephyr_discovery.KubernetesClusterClient,
	localSecretClient k8s_core_v1_clients.SecretClient,
	secretClientFactory k8s_core_v1_clients.SecretClientFactory,
	dynamicClientGetter multicluster.DynamicClientGetter,
	serviceAccountClientFactory k8s_core_v1_clients.ServiceAccountClientFactory,
) ClusterDeregistrationClient {
	return &clusterDeregistrationClient{
		crdRemover:                  crdRemover,
		csrAgentInstallerFactory:    csrAgentInstallerFactory,
		kubeConfigLookup:            kubeConfigLookup,
		localKubeClusterClient:      localKubeClusterClient,
		localSecretClient:           localSecretClient,
		secretClientFactory:         secretClientFactory,
		dynamicClientGetter:         dynamicClientGetter,
		serviceAccountClientFactory: serviceAccountClientFactory,
	}
}

type clusterDeregistrationClient struct {
	crdRemover                  crd_uninstall.CrdRemover
	kubeLoader                  kubeconfig.KubeLoader
	csrAgentInstallerFactory    installation.CsrAgentInstallerFactory
	kubeConfigLookup            kubeconfig.KubeConfigLookup
	localKubeClusterClient      zephyr_discovery.KubernetesClusterClient
	localSecretClient           k8s_core_v1_clients.SecretClient
	secretClientFactory         k8s_core_v1_clients.SecretClientFactory
	serviceAccountClientFactory k8s_core_v1_clients.ServiceAccountClientFactory
	dynamicClientGetter         multicluster.DynamicClientGetter
}

func (c *clusterDeregistrationClient) Deregister(ctx context.Context, kubeCluster *zephyr_discovery.KubernetesCluster) error {
	config, err := c.kubeConfigLookup.FromCluster(ctx, kubeCluster.GetName())
	if meta.IsNoMatchError(err) {
		return nil
	} else if err != nil {
		return FailedToFindClusterCredentials(err, kubeCluster.GetName())
	}
	kubeClient := kubernetes.NewForConfigOrDie(config.RestConfig)
	helmInstallerFactory := helm.NewHelmInstallerFactory(kubeClient.CoreV1().Namespaces(), ioutil.Discard)
	csrAgentInstaller := c.csrAgentInstallerFactory(helmInstallerFactory)
	err = csrAgentInstaller.Uninstall(&installation.CsrAgentUninstallOptions{
		KubeConfig:       installation.KubeConfig{KubeConfig: config.ClientConfig},
		ReleaseName:      constants.CsrAgentReleaseName,
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

	err = c.localKubeClusterClient.DeleteKubernetesCluster(ctx, selection.ObjectMetaToObjectKey(kubeCluster.ObjectMeta))
	if err != nil {
		return FailedToCleanUpKubeConfigCrd(err, kubeCluster.GetName())
	}

	// the CSR agent installs only CRDs from the security group. Remove only those
	_, err = c.crdRemover.RemoveCrdGroup(ctx, kubeCluster.GetName(), config.RestConfig, zephyr_security_scheme.SchemeGroupVersion)
	if err != nil && !meta.IsNoMatchError(err) {
		return FailedToRemoveCrds(err, kubeCluster.GetName())
	}

	// Remove the service account as the last step, as this is required for SMH to perform any operation on the cluster.
	err = c.cleanUpServiceAccounts(ctx, clientForCluster, kubeCluster)
	if err != nil {
		return FailedToCleanUpServiceAccount(err, kubeCluster.GetName())
	}

	return nil
}

func (c *clusterDeregistrationClient) cleanUpServiceAccounts(ctx context.Context, clientForCluster client.Client, kubeCluster *zephyr_discovery.KubernetesCluster) error {
	serviceAccountClient := c.serviceAccountClientFactory(clientForCluster)
	err := serviceAccountClient.DeleteAllOfServiceAccount(
		ctx,
		client.InNamespace(kubeCluster.Spec.GetWriteNamespace()),
		client.MatchingLabels{
			constants.ManagedByLabel:        constants.ServiceMeshHubApplicationName,
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
			err := secretClientForCluster.DeleteSecret(ctx, selection.ObjectMetaToObjectKey(secret.ObjectMeta))

			// if we have multiple de-registrations going on at once (potentially in `meshctl uninstall`, ignore the error if something else cleaned up the secret first)
			if err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (c *clusterDeregistrationClient) cleanUpKubeConfigSecret(ctx context.Context, kubeCluster *zephyr_discovery.KubernetesCluster) error {
	kubeConfigSecret, err := c.localSecretClient.GetSecret(ctx, selection.ResourceRefToObjectKey(kubeCluster.Spec.GetSecretRef()))
	if err != nil {
		return err
	}

	err = c.localSecretClient.DeleteSecret(ctx, selection.ObjectMetaToObjectKey(kubeConfigSecret.ObjectMeta))
	if err != nil {
		return err
	}

	return nil
}
