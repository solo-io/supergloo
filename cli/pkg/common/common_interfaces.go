package common

import (
	"io/ioutil"
	"os"

	"github.com/solo-io/go-utils/installutils/helminstall/types"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/common/kube"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	healthcheck_types "github.com/solo-io/mesh-projects/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/check/status"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/deregister"
	register "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register/csr"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator"
	crd_uninstall "github.com/solo-io/mesh-projects/cli/pkg/tree/uninstall/crd"
	upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/auth"
	kubernetes_apiext "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/apiext"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/kubeconfig"
	"github.com/solo-io/mesh-projects/pkg/version"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination ../mocks/mock_common_interfaces.go -package cli_mocks -source ./common_interfaces.go

// a grab bag of various clients that command implementations may use
type KubeClients struct {
	ClusterAuthorization            auth.ClusterAuthorization
	SecretClient                    kubernetes_core.SecretsClient
	HelmInstaller                   types.Installer
	HelmClient                      types.HelmClient                       // used for uninstalling - the go-utils package is not laid out very well
	KubeClusterClient               discovery_core.KubernetesClusterClient // client for KubernetesCluster custom resources
	DeployedVersionFinder           version.DeployedVersionFinder
	CrdClientFactory                kubernetes_apiext.GeneratedCrdClientFactory
	HealthCheckClients              healthcheck_types.Clients
	SecretsClient                   kubernetes_core.SecretsClient
	NamespaceClient                 kubernetes_core.NamespaceClient
	UninstallClients                UninstallClients
	InMemoryRESTClientGetterFactory common_config.InMemoryRESTClientGetterFactory
	ClusterDeregistrationClient     deregister.ClusterDeregistrationClient
}

type KubeClientsFactory func(masterConfig *rest.Config, writeNamespace string) (*KubeClients, error)

type Clients struct {
	ServerVersionClient           server.ServerVersionClient
	MasterClusterVerifier         common_config.MasterKubeConfigVerifier
	ReleaseAssetHelper            upgrade_assets.AssetHelper
	UnstructuredKubeClientFactory kube.UnstructuredKubeClientFactory
	DeploymentClient              server.DeploymentClient
	StatusClientFactory           status.StatusClientFactory
	HealthCheckSuite              healthcheck_types.HealthCheckSuite

	IstioClients               IstioClients
	ClusterRegistrationClients ClusterRegistrationClients
}

func IstioClientsProvider(manifestBuilder operator.InstallerManifestBuilder, operatorManagerFactory operator.OperatorManagerFactory) IstioClients {
	return IstioClients{
		OperatorManifestBuilder: manifestBuilder,
		OperatorManagerFactory:  operatorManagerFactory,
	}
}

type IstioClients struct {
	OperatorManifestBuilder operator.InstallerManifestBuilder
	OperatorManagerFactory  operator.OperatorManagerFactory
}

type UninstallClients struct {
	CrdRemover              crd_uninstall.CrdRemover
	SecretToConfigConverter kubeconfig.SecretToConfigConverter
}

func UninstallClientsProvider(
	crdRemover crd_uninstall.CrdRemover,
	secretToConfigConverter kubeconfig.SecretToConfigConverter,
) UninstallClients {
	return UninstallClients{
		CrdRemover:              crdRemover,
		SecretToConfigConverter: secretToConfigConverter,
	}
}

func ClusterRegistrationClientsProvider(csrAgentInstallerFactory register.CsrAgentInstallerFactory) ClusterRegistrationClients {
	return ClusterRegistrationClients{
		CsrAgentInstallerFactory: csrAgentInstallerFactory,
	}
}

type ClusterRegistrationClients struct {
	CsrAgentInstallerFactory register.CsrAgentInstallerFactory
}

type ClientsFactory func(opts *options.Options) (*Clients, error)

func ClientsProvider(
	serverVersionClient server.ServerVersionClient,
	assetHelper upgrade_assets.AssetHelper,
	verifier common_config.MasterKubeConfigVerifier,
	unstructuredKubeClientFactory kube.UnstructuredKubeClientFactory,
	deploymentClient server.DeploymentClient,
	istioClients IstioClients,
	statusClientFactory status.StatusClientFactory,
	healthCheckSuite healthcheck_types.HealthCheckSuite,
	clusterRegistrationClients ClusterRegistrationClients,
) *Clients {
	return &Clients{
		ServerVersionClient:           serverVersionClient,
		MasterClusterVerifier:         verifier,
		UnstructuredKubeClientFactory: unstructuredKubeClientFactory,
		DeploymentClient:              deploymentClient,
		ReleaseAssetHelper:            assetHelper,
		IstioClients:                  istioClients,
		StatusClientFactory:           statusClientFactory,
		HealthCheckSuite:              healthCheckSuite,
		ClusterRegistrationClients:    clusterRegistrationClients,
	}
}

// facilitates wire codegen
func KubeClientsProvider(
	authorization auth.ClusterAuthorization,
	helmInstaller types.Installer,
	helmClient types.HelmClient,
	kubeClusterClient discovery_core.KubernetesClusterClient,
	healthCheckClients healthcheck_types.Clients,
	deployedVersionFinder version.DeployedVersionFinder,
	crdClientFactory kubernetes_apiext.GeneratedCrdClientFactory,
	secretsClient kubernetes_core.SecretsClient,
	namespaceClient kubernetes_core.NamespaceClient,
	uninstallClients UninstallClients,
	inMemoryRESTClientGetterFactory common_config.InMemoryRESTClientGetterFactory,
	clusterDeregistrationClient deregister.ClusterDeregistrationClient,
) *KubeClients {
	return &KubeClients{
		ClusterAuthorization:            authorization,
		HelmInstaller:                   helmInstaller,
		HelmClient:                      helmClient,
		KubeClusterClient:               kubeClusterClient,
		DeployedVersionFinder:           deployedVersionFinder,
		CrdClientFactory:                crdClientFactory,
		HealthCheckClients:              healthCheckClients,
		SecretsClient:                   secretsClient,
		NamespaceClient:                 namespaceClient,
		UninstallClients:                uninstallClients,
		InMemoryRESTClientGetterFactory: inMemoryRESTClientGetterFactory,
		ClusterDeregistrationClient:     clusterDeregistrationClient,
	}
}

type FileReader interface {
	Read(filePath string) ([]byte, error)
	Exists(filePath string) (exists bool, err error)
}

func NewDefaultFileReader() FileReader {
	return &fileReader{}
}

type fileReader struct{}

func (f *fileReader) Read(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func (f *fileReader) Exists(filePath string) (exists bool, err error) {
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
