package cluster_registration

import (
	"context"
	"fmt"
	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	"io/ioutil"
	"net"
	"net/url"
	"strings"

	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	constants2 "github.com/solo-io/service-mesh-hub/pkg/common/constants"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/installation"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/auth"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/helm"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
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
	ContextNotFound = func(contextName string) error {
		return eris.Errorf("Context with name %s not found in kubeconfig", contextName)
	}
	ClusterNotFound = func(clusterName string) error {
		return eris.Errorf("Cluster with name %s not found in kubeconfig", clusterName)
	}
)

type clusterRegistrationClient struct {
	secretClient                       k8s_core.SecretClient
	kubernetesClusterClient            smh_discovery.KubernetesClusterClient
	namespaceClientFactory             k8s_core_providers.NamespaceClientFromConfigFactory
	kubeConverter                      kubeconfig.Converter
	csrAgentInstallerFactory           installation.CsrAgentInstallerFactory
	clusterAuthClientFromConfigFactory auth.ClusterAuthClientFromConfigFactory
}

func NewClusterRegistrationClient(
	secretClient k8s_core.SecretClient,
	kubernetesClusterClient smh_discovery.KubernetesClusterClient,
	namespaceClientFactory k8s_core_providers.NamespaceClientFromConfigFactory,
	kubeConverter kubeconfig.Converter,
	csrAgentInstallerFactory installation.CsrAgentInstallerFactory,
	clusterAuthClientFromConfigFactory auth.ClusterAuthClientFromConfigFactory,
) ClusterRegistrationClient {
	return &clusterRegistrationClient{
		secretClient:                       secretClient,
		kubernetesClusterClient:            kubernetesClusterClient,
		namespaceClientFactory:             namespaceClientFactory,
		kubeConverter:                      kubeConverter,
		csrAgentInstallerFactory:           csrAgentInstallerFactory,
		clusterAuthClientFromConfigFactory: clusterAuthClientFromConfigFactory,
	}
}

type ClusterRegisterOpts struct {
	Overwrite                        bool
	UseDevCsrAgentChart              bool
	LocalClusterDomainOverride       string
	CsrAgentHelmChartValuesFileNames map[string]interface{}
}

func (c *clusterRegistrationClient) Register(
	ctx context.Context,
	remoteConfig clientcmd.ClientConfig,
	remoteClusterName string,
	remoteWriteNamespace string,
	remoteContextName string,
	discoverySource string,
	clusterRegisterOpts ClusterRegisterOpts,
) error {
	remoteRestConfig, err := remoteConfig.ClientConfig()
	if err != nil {
		return err
	}
	err = c.checkClusterExistence(ctx, remoteClusterName, clusterRegisterOpts.Overwrite)
	if err != nil {
		return err
	}
	err = c.ensureRemoteNamespace(ctx, remoteRestConfig, remoteWriteNamespace)
	if err != nil {
		return err
	}
	serviceAccountBearerToken, err := c.generateServiceAccountBearerToken(
		ctx,
		remoteRestConfig,
		remoteClusterName,
		remoteWriteNamespace,
	)
	if err != nil {
		return err
	}
	// Install CRDs to remote cluster. This must happen before kubeconfig Secret is written to master cluster because
	// relevant CRDs must exist before SMH attempts any cross cluster functionality.
	err = c.installRemoteCSRAgent(
		ctx,
		remoteWriteNamespace,
		remoteConfig,
		clusterRegisterOpts.UseDevCsrAgentChart,
		clusterRegisterOpts.CsrAgentHelmChartValuesFileNames,
	)
	if err != nil {
		return err
	}
	// Write kubeconfig Secret and KubeCluster CRD to master cluster
	secret, err := c.writeKubeConfigToMaster(
		ctx,
		remoteClusterName,
		remoteContextName,
		serviceAccountBearerToken,
		remoteConfig,
		clusterRegisterOpts.LocalClusterDomainOverride,
	)
	if err != nil {
		return err
	}
	err = c.writeKubeClusterToMaster(
		ctx,
		container_runtime.GetWriteNamespace(),
		remoteClusterName,
		remoteWriteNamespace,
		secret,
		discoverySource,
	)
	if err != nil {
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
			Namespace: container_runtime.GetWriteNamespace(),
		},
	)
	if k8s_errs.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	} else if !overwrite {
		return ClusterAlreadyRegisteredError(remoteClusterName)
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
	_, err = remoteNamespaceClient.GetNamespace(ctx, writeNamespace)
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

func (c *clusterRegistrationClient) installRemoteCSRAgent(
	ctx context.Context,
	remoteWriteNamespace string,
	remoteConfig clientcmd.ClientConfig,
	useDevCsrAgentChart bool,
	csrAgentHelmValues map[string]interface{},
) error {
	restConfig, err := remoteConfig.ClientConfig()
	if err != nil {
		return err
	}
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	helmInstaller := helm.NewHelmInstallerFactory(kubeClient.CoreV1().Namespaces(), ioutil.Discard)
	csrAgentInstaller := c.csrAgentInstallerFactory(helmInstaller)
	return csrAgentInstaller.Install(
		ctx,
		&installation.CsrAgentInstallOptions{
			KubeConfig:           installation.KubeConfig{KubeConfig: remoteConfig},
			SmhInstallNamespace:  container_runtime.GetWriteNamespace(),
			UseDevCsrAgentChart:  useDevCsrAgentChart,
			ReleaseName:          constants2.CsrAgentReleaseName,
			RemoteWriteNamespace: remoteWriteNamespace,
			ExtraValues:          csrAgentHelmValues,
		})
}

func (c *clusterRegistrationClient) generateServiceAccountBearerToken(
	ctx context.Context,
	remoteAuthConfig *rest.Config,
	remoteClusterName string,
	remoteWriteNamespace string,
) (string, error) {
	serviceAccountRef := &smh_core_types.ResourceRef{
		Name:      remoteClusterName,
		Namespace: remoteWriteNamespace,
	}
	remoteClusterAuth, err := c.clusterAuthClientFromConfigFactory(remoteAuthConfig)
	if err != nil {
		return "", err
	}
	bearerTokenForServiceAccount, err := remoteClusterAuth.BuildRemoteBearerToken(ctx, remoteAuthConfig, serviceAccountRef)
	if err != nil {
		return "", err
	}
	return bearerTokenForServiceAccount, nil
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
	remoteContext, ok := config.Contexts[remoteContextName]
	if !ok {
		return nil, ContextNotFound(remoteContextName)
	}
	remoteCluster, ok := config.Clusters[remoteContext.Cluster]
	if !ok {
		return nil, ClusterNotFound(remoteContext.Cluster)
	}
	// Hack for local e2e testing with Kind
	err = hackClusterConfigForLocalTestingInKIND(remoteCluster, remoteContextName, localClusterDomainOverride)
	if err != nil {
		return nil, err
	}
	secret, err := c.kubeConverter.ConfigToSecret(
		remoteClusterName,
		container_runtime.GetWriteNamespace(),
		&kubeconfig.KubeConfig{
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
	discoverySource string,
) error {
	cluster := &smh_discovery.KubernetesCluster{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      remoteClusterName,
			Namespace: writeNamespace,
			Labels: map[string]string{
				kube.DISCOVERED_BY: discoverySource,
			},
		},
		Spec: smh_discovery_types.KubernetesClusterSpec{
			SecretRef: &smh_core_types.ResourceRef{
				Name:      secret.GetName(),
				Namespace: secret.GetNamespace(),
			},
			WriteNamespace: remoteWriteNamespace,
		},
	}
	err := c.kubernetesClusterClient.UpsertKubernetesCluster(ctx, cluster)
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
	if k8s_errs.IsNotFound(err) {
		return c.secretClient.CreateSecret(ctx, secret)
	} else if err != nil {
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
		port := serverUrl.Port()
		if host, newPort, err := net.SplitHostPort(clusterDomainOverride); err == nil {
			clusterDomainOverride = host
			port = newPort
		}
		remoteCluster.Server = fmt.Sprintf("https://%s:%s", clusterDomainOverride, port)
		remoteCluster.InsecureSkipTLSVerify = true
		remoteCluster.CertificateAuthority = ""
		remoteCluster.CertificateAuthorityData = []byte("")
	}
	return nil
}
