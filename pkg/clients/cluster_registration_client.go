package clients

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register/csr"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/version"
	"go.uber.org/zap/zaptest"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	FailedToWriteSecret = func(err error) error {
		return eris.Wrap(err, "Could not write secret to master cluster")
	}
	FailedToEnsureWriteNamespace = func(err error, namespace string) error {
		return eris.Wrapf(err, "Failed to find or create namespace %s on the cluster being registered", namespace)
	}
	ClusterAlreadyRegisteredError = func(remoteClusterName string) error {
		return eris.Errorf("Cluster %s already registered.", remoteClusterName)
	}
	FailedToWriteKubeCluster = func(err error) error {
		return eris.Wrap(err, "Failed to write KubernetesCluster resource to master cluster")
	}
	FailedToConvertToSecret = func(err error) error {
		return eris.Wrap(err, "Could not convert kube config for new service account into secret")
	}
	EmptyContextsError = eris.New("No contexts found for kube config")
)

type clusterRegistrationClient struct {
	secretClient                        k8s_core.SecretClient
	namespaceClient                     k8s_core.NamespaceClient
	kubernetesClusterClient             zephyr_discovery.KubernetesClusterClient
	helmClientFactory                   types.HelmClientForMemoryConfigFactory
	deployedVersionFinder               version.DeployedVersionFinder
	kubeConverter                       kube.Converter
	namespaceClientFactory              k8s_core.NamespaceClientFromConfigFactory
	secretClientFactory                 k8s_core.SecretClientFromConfigFactory
	serviceAccountClientFactory         k8s_core.ServiceAccountClientFromConfigFactory
	rbacClientFactory                   auth.RbacClientFactory
	remoteAuthorityConfigCreatorFactory auth.RemoteAuthorityConfigCreatorFactory
	remoteAuthorityManagerFactory       auth.RemoteAuthorityManagerFactory
	clusterAuthorizationFactory         auth.ClusterAuthorizationFactory
	csrAgentInstallerFactory            csr.CsrAgentInstallerFactory
}

func NewClusterRegistrationClient(
	secretClient k8s_core.SecretClient,
	namespaceClient k8s_core.NamespaceClient,
	kubernetesClusterClient zephyr_discovery.KubernetesClusterClient,
	helmClientFactory types.HelmClientForMemoryConfigFactory,
	deployedVersionFinder version.DeployedVersionFinder,
	kubeConverter kube.Converter,
	namespaceClientFactory k8s_core.NamespaceClientFromConfigFactory,
	secretClientFactory k8s_core.SecretClientFromConfigFactory,
	serviceAccountClientFactory k8s_core.ServiceAccountClientFromConfigFactory,
	rbacClientFactory auth.RbacClientFactory,
	remoteAuthorityConfigCreatorFactory auth.RemoteAuthorityConfigCreatorFactory,
	remoteAuthorityManagerFactory auth.RemoteAuthorityManagerFactory,
	clusterAuthorizationFactory auth.ClusterAuthorizationFactory,
	csrAgentInstallerFactory csr.CsrAgentInstallerFactory,
) ClusterRegistrationClient {
	return &clusterRegistrationClient{
		secretClient:                        secretClient,
		namespaceClient:                     namespaceClient,
		kubernetesClusterClient:             kubernetesClusterClient,
		helmClientFactory:                   helmClientFactory,
		deployedVersionFinder:               deployedVersionFinder,
		kubeConverter:                       kubeConverter,
		namespaceClientFactory:              namespaceClientFactory,
		secretClientFactory:                 secretClientFactory,
		serviceAccountClientFactory:         serviceAccountClientFactory,
		rbacClientFactory:                   rbacClientFactory,
		remoteAuthorityConfigCreatorFactory: remoteAuthorityConfigCreatorFactory,
		remoteAuthorityManagerFactory:       remoteAuthorityManagerFactory,
		clusterAuthorizationFactory:         clusterAuthorizationFactory,
		csrAgentInstallerFactory:            csrAgentInstallerFactory,
	}
}

func (c *clusterRegistrationClient) Register(
	ctx context.Context,
	remoteConfig clientcmd.ClientConfig,
	remoteClusterName string,
	remoteWriteNamespace string,
	overwrite bool,
	useDevCsrAgentChart bool,
	localClusterDomainOverride string,
	remoteContextName string,
) error {
	var err error
	var remoteRestConfig *rest.Config
	var serviceAccountBearerToken string
	var secret *k8s_core_types.Secret
	if remoteRestConfig, err = remoteConfig.ClientConfig(); err != nil {
		return err
	}
	if err = c.checkClusterExistence(ctx, remoteClusterName, overwrite); err != nil {
		return err
	}
	if err = c.ensureRemoteNamespace(ctx, remoteRestConfig, remoteWriteNamespace); err != nil {
		return err
	}
	if serviceAccountBearerToken, err = c.generateServiceAccountBearerToken(
		ctx,
		remoteRestConfig,
		remoteClusterName,
		remoteWriteNamespace,
	); err != nil {
		return err
	}
	// Install CRDs to remote cluster. This must happen before kubeconfig Secret is written to master cluster because
	// relevant CRDs must exist before SMH attempts any cross cluster functionality.
	if err = c.installRemoteCRDs(ctx, remoteClusterName, remoteWriteNamespace, remoteConfig, useDevCsrAgentChart); err != nil {
		return err
	}
	// Write kubeconfig Secret and KubeCluster CRD to master cluster
	if secret, err = c.writeKubeConfigToMaster(
		ctx,
		remoteClusterName,
		remoteContextName,
		serviceAccountBearerToken,
		remoteConfig,
		localClusterDomainOverride,
	); err != nil {
		return err
	}
	if err = c.writeKubeClusterToMaster(ctx, env.GetWriteNamespace(), remoteClusterName, remoteWriteNamespace, secret); err != nil {
		return err
	}
	return nil
}

func (c *clusterRegistrationClient) checkClusterExistence(
	ctx context.Context,
	remoteClusterName string,
	overwrite bool,
) error {
	_, err := c.kubernetesClusterClient.GetKubernetesCluster(
		ctx,
		client.ObjectKey{
			Name:      remoteClusterName,
			Namespace: env.GetWriteNamespace(),
		},
	)
	if err != nil {
		if !k8s_errs.IsNotFound(err) {
			return err
		}
	} else {
		if !overwrite {
			return ClusterAlreadyRegisteredError(remoteClusterName)
		}
	}
	return nil
}

func (c *clusterRegistrationClient) ensureRemoteNamespace(
	ctx context.Context,
	remoteRestConfig *rest.Config,
	writeNamespace string,
) error {
	remoteNamespaceClient, err := c.namespaceClientFactory(remoteRestConfig)
	if err != nil {
		return err
	}
	_, err = remoteNamespaceClient.GetNamespace(ctx, client.ObjectKey{Name: writeNamespace})
	if k8s_errs.IsNotFound(err) {
		return remoteNamespaceClient.CreateNamespace(ctx, &k8s_core_types.Namespace{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: writeNamespace,
			},
		})
	} else if err != nil {
		return FailedToEnsureWriteNamespace(err, writeNamespace)
	}
	return nil
}

func (c *clusterRegistrationClient) installRemoteCRDs(
	ctx context.Context,
	remoteClusterName string,
	remoteWriteNamespace string,
	remoteConfig clientcmd.ClientConfig,
	useDevCsrAgentChart bool,
) error {
	restConfig, err := remoteConfig.ClientConfig()
	if err != nil {
		return err
	}
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	helmInstaller := helminstall.NewInstallerFactory(kubeClient.CoreV1().Namespaces(), &zaptest.Discarder{})
	csrAgentInstaller := c.csrAgentInstallerFactory(helmInstaller)
	return csrAgentInstaller.Install(
		ctx,
		&csr.CsrAgentInstallOptions{
			KubeConfig:           remoteConfig,
			ClusterName:          remoteClusterName,
			SmhInstallNamespace:  env.GetWriteNamespace(),
			UseDevCsrAgentChart:  useDevCsrAgentChart,
			ReleaseName:          cliconstants.CsrAgentReleaseName,
			RemoteWriteNamespace: remoteWriteNamespace,
		})
}

func (c *clusterRegistrationClient) generateServiceAccountBearerToken(
	ctx context.Context,
	remoteAuthConfig *rest.Config,
	remoteClusterName string,
	remoteWriteNamespace string,
) (string, error) {
	serviceAccountRef := &zephyr_core_types.ResourceRef{
		Name:      remoteClusterName,
		Namespace: remoteWriteNamespace,
	}
	remoteClusterAuth, err := c.constructRemoteClusterAuth(remoteAuthConfig)
	if err != nil {
		return "", err
	}
	bearerTokenForServiceAccount, err := remoteClusterAuth.BuildRemoteBearerToken(ctx, remoteAuthConfig, serviceAccountRef)
	if err != nil {
		return "", err
	}
	return bearerTokenForServiceAccount, nil
}

func (c *clusterRegistrationClient) constructRemoteClusterAuth(remoteAuthConfig *rest.Config) (auth.ClusterAuthorization, error) {
	serviceAccountClient, err := c.serviceAccountClientFactory(remoteAuthConfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(remoteAuthConfig)
	if err != nil {
		return nil, err
	}
	rbacClient := c.rbacClientFactory(clientset)
	secretClient, err := c.secretClientFactory(remoteAuthConfig)
	if err != nil {
		return nil, err
	}
	remoteAuthorityConfigCreator := c.remoteAuthorityConfigCreatorFactory(secretClient, serviceAccountClient)
	remoteAuthorityManager := c.remoteAuthorityManagerFactory(serviceAccountClient, rbacClient)
	return c.clusterAuthorizationFactory(remoteAuthorityConfigCreator, remoteAuthorityManager), nil
}

func (c *clusterRegistrationClient) writeKubeConfigToMaster(
	ctx context.Context,
	remoteClusterName string,
	remoteContextName string,
	serviceAccountBearerToken string,
	clientConfig clientcmd.ClientConfig,
	localClusterDomainOverride string,
) (*k8s_core_types.Secret, error) {
	config, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}
	if len(config.Contexts) < 1 {
		return nil, EmptyContextsError
	}
	var remoteContext *api.Context
	if len(config.Contexts) == 1 {
		for _, context := range config.Contexts {
			remoteContext = context
		}
	} else {
		remoteContext = config.Contexts[remoteContextName]
	}
	remoteCluster := config.Clusters[remoteContext.Cluster]
	// Hack for local e2e testing with Kind
	err = hackClusterConfigForLocalTestingInKIND(remoteCluster, remoteContextName, localClusterDomainOverride)
	if err != nil {
		return nil, err
	}
	secret, err := c.kubeConverter.ConfigToSecret(
		remoteClusterName,
		env.GetWriteNamespace(),
		&kube.KubeConfig{
			Config: api.Config{
				Kind:        "Secret",
				APIVersion:  "kubernetes_core",
				Preferences: api.Preferences{},
				Clusters: map[string]*api.Cluster{
					remoteClusterName: remoteCluster,
				},
				AuthInfos: map[string]*api.AuthInfo{
					remoteClusterName: {
						Token: serviceAccountBearerToken,
					},
				},
				Contexts: map[string]*api.Context{
					remoteClusterName: {
						LocationOfOrigin: remoteContext.LocationOfOrigin,
						Cluster:          remoteClusterName,
						AuthInfo:         remoteClusterName,
						Namespace:        remoteContext.Namespace,
						Extensions:       remoteContext.Extensions,
					},
				},
				CurrentContext: remoteClusterName,
			},
			Cluster: remoteClusterName,
		})
	if err != nil {
		return nil, FailedToConvertToSecret(err)
	}
	err = c.upsertSecretData(ctx, secret)
	if err != nil {
		return nil, FailedToWriteSecret(err)
	}

	return secret, nil
}

func (c *clusterRegistrationClient) writeKubeClusterToMaster(
	ctx context.Context,
	writeNamespace string,
	remoteClusterName string,
	remoteWriteNamespace string,
	secret *k8s_core_types.Secret,
) error {
	cluster := &zephyr_discovery.KubernetesCluster{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      remoteClusterName,
			Namespace: writeNamespace,
		},
		Spec: zephyr_discovery_types.KubernetesClusterSpec{
			SecretRef: &zephyr_core_types.ResourceRef{
				Name:      secret.GetName(),
				Namespace: secret.GetNamespace(),
			},
			WriteNamespace: remoteWriteNamespace,
		},
	}
	err := c.kubernetesClusterClient.UpsertKubernetesClusterSpec(ctx, cluster)
	if err != nil {
		return FailedToWriteKubeCluster(err)
	}
	return nil
}

func (c *clusterRegistrationClient) upsertSecretData(
	ctx context.Context,
	secret *k8s_core_types.Secret,
) error {
	existing, err := c.secretClient.GetSecret(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace})
	if err != nil {
		if k8s_errs.IsNotFound(err) {
			return c.secretClient.CreateSecret(ctx, secret)
		}
		return err
	}
	existing.Data = secret.Data
	existing.StringData = secret.StringData
	return c.secretClient.UpdateSecret(ctx, existing)
}

// if:
//   * we are operating against a context named "kind-", AND
//   * the server appears to point to localhost, AND
//   * the --local-cluster-domain-override flag is populated with a value
//
// then we rewrite the server config to communicate over the value of `--local-cluster-domain-override`, which
// resolves to the host machine of docker. We also need to skip TLS verification
// and zero-out the cert data, because the cert on the remote cluster's API server wasn't
// issued for the domain contained in the value of `--local-cluster-domain-override`.
//
// this function call is a no-op if those conditions are not met
func hackClusterConfigForLocalTestingInKIND(
	remoteCluster *api.Cluster,
	remoteContextName, clusterDomainOverride string,
) error {
	serverUrl, err := url.Parse(remoteCluster.Server)
	if err != nil {
		return err
	}
	if strings.HasPrefix(remoteContextName, "kind-") &&
		(serverUrl.Hostname() == "127.0.0.1" || serverUrl.Hostname() == "localhost") &&
		clusterDomainOverride != "" {
		remoteCluster.Server = fmt.Sprintf("https://%s:%s", clusterDomainOverride, serverUrl.Port())
		remoteCluster.InsecureSkipTLSVerify = true
		remoteCluster.CertificateAuthority = ""
		remoteCluster.CertificateAuthorityData = []byte("")
	}
	return nil
}
