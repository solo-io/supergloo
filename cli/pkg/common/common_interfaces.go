package common

import (
	"io/ioutil"
	"os"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/status"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/deregister"
	register "github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register/csr"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/describe/description"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio/operator"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/config_lookup"
	crd_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/crd"
	upgrade_assets "github.com/solo-io/service-mesh-hub/cli/pkg/tree/upgrade/assets"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/version/server"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apiextensions "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apiextensions.k8s.io/v1beta1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	"github.com/solo-io/service-mesh-hub/pkg/version"
	"k8s.io/client-go/rest"
)

var (
	FailedLoadingMasterConfig = func(err error) error {
		return eris.Wrap(err, "Failed to load the kube config for the master cluster")
	}
)

//go:generate mockgen -destination ../mocks/mock_common_interfaces.go -package cli_mocks -source ./common_interfaces.go

// a grab bag of various clients that command implementations may use
type KubeClients struct {
	ClusterAuthorization            auth.ClusterAuthorization
	HelmInstaller                   types.Installer
	HelmClient                      types.HelmClient                         // used for uninstalling - the go-utils package is not laid out very well
	KubeClusterClient               zephyr_discovery.KubernetesClusterClient // client for KubernetesCluster custom resources
	MeshServiceClient               zephyr_discovery.MeshServiceClient
	MeshWorkloadClient              zephyr_discovery.MeshWorkloadClient
	MeshClient                      zephyr_discovery.MeshClient
	VirtualMeshClient               zephyr_networking.VirtualMeshClient
	VirtualMeshCSRClient            zephyr_security.VirtualMeshCertificateSigningRequestClient
	DeployedVersionFinder           version.DeployedVersionFinder
	CrdClientFactory                k8s_apiextensions.CustomResourceDefinitionClientFromConfigFactory
	HealthCheckClients              healthcheck_types.Clients
	SecretClient                    k8s_core.SecretClient
	NamespaceClient                 k8s_core.NamespaceClient
	UninstallClients                UninstallClients
	InMemoryRESTClientGetterFactory common_config.InMemoryRESTClientGetterFactory
	ClusterDeregistrationClient     deregister.ClusterDeregistrationClient
	KubeConfigLookup                config_lookup.KubeConfigLookup
	TrafficPolicyClient             zephyr_networking.TrafficPolicyClient
	AccessControlPolicyClient       zephyr_networking.AccessControlPolicyClient
	ResourceDescriber               description.ResourceDescriber
	ResourceSelector                selector.ResourceSelector
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

func IstioClientsProvider(
	manifestBuilder operator.InstallerManifestBuilder,
	operatorManagerFactory operator.OperatorManagerFactory,
) IstioClients {
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

func ClusterRegistrationClientsProvider(
	csrAgentInstallerFactory register.CsrAgentInstallerFactory,
) ClusterRegistrationClients {
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
	kubeClusterClient zephyr_discovery.KubernetesClusterClient,
	healthCheckClients healthcheck_types.Clients,
	deployedVersionFinder version.DeployedVersionFinder,
	crdClientFactory k8s_apiextensions.CustomResourceDefinitionClientFromConfigFactory,
	secretClient k8s_core.SecretClient,
	namespaceClient k8s_core.NamespaceClient,
	uninstallClients UninstallClients,
	inMemoryRESTClientGetterFactory common_config.InMemoryRESTClientGetterFactory,
	clusterDeregistrationClient deregister.ClusterDeregistrationClient,
	kubeConfigLookup config_lookup.KubeConfigLookup,
	virtualMeshCsrClient zephyr_security.VirtualMeshCertificateSigningRequestClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshClient zephyr_discovery.MeshClient,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
	resourceDescriber description.ResourceDescriber,
	resourceSelector selector.ResourceSelector,
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	accessControlPolicyClient zephyr_networking.AccessControlPolicyClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
) *KubeClients {
	return &KubeClients{
		ClusterAuthorization:            authorization,
		HelmInstaller:                   helmInstaller,
		HelmClient:                      helmClient,
		KubeClusterClient:               kubeClusterClient,
		MeshServiceClient:               meshServiceClient,
		VirtualMeshCSRClient:            virtualMeshCsrClient,
		DeployedVersionFinder:           deployedVersionFinder,
		CrdClientFactory:                crdClientFactory,
		HealthCheckClients:              healthCheckClients,
		SecretClient:                    secretClient,
		NamespaceClient:                 namespaceClient,
		UninstallClients:                uninstallClients,
		InMemoryRESTClientGetterFactory: inMemoryRESTClientGetterFactory,
		ClusterDeregistrationClient:     clusterDeregistrationClient,
		KubeConfigLookup:                kubeConfigLookup,
		MeshClient:                      meshClient,
		VirtualMeshClient:               virtualMeshClient,
		ResourceDescriber:               resourceDescriber,
		ResourceSelector:                resourceSelector,
		TrafficPolicyClient:             trafficPolicyClient,
		AccessControlPolicyClient:       accessControlPolicyClient,
		MeshWorkloadClient:              meshWorkloadClient,
	}
}

type Printers struct {
	MeshPrinter                table_printing.MeshPrinter
	MeshServicePrinter         table_printing.MeshServicePrinter
	MeshWorkloadPrinter        table_printing.MeshWorkloadPrinter
	KubeClusterPrinter         table_printing.KubernetesClusterPrinter
	VirtualMeshPrinter         table_printing.VirtualMeshPrinter
	VirtualMeshCSRPrinter      table_printing.VirtualMeshCSRPrinter
	TrafficPolicyPrinter       table_printing.TrafficPolicyPrinter
	AccessControlPolicyPrinter table_printing.AccessControlPolicyPrinter
	ResourcePrinter            resource_printing.ResourcePrinter
}

func PrintersProvider(
	meshPrinter table_printing.MeshPrinter,
	meshServicePrinter table_printing.MeshServicePrinter,
	meshWorkloadPrinter table_printing.MeshWorkloadPrinter,
	kubeClusterPrinter table_printing.KubernetesClusterPrinter,
	trafficPolicyPrinter table_printing.TrafficPolicyPrinter,
	accessControlPolicyPrinter table_printing.AccessControlPolicyPrinter,
	virtualMeshPrinter table_printing.VirtualMeshPrinter,
	vmcsrPrinter table_printing.VirtualMeshCSRPrinter,
	resourcePrinter resource_printing.ResourcePrinter,
) Printers {
	return Printers{
		MeshPrinter:                meshPrinter,
		MeshServicePrinter:         meshServicePrinter,
		MeshWorkloadPrinter:        meshWorkloadPrinter,
		KubeClusterPrinter:         kubeClusterPrinter,
		TrafficPolicyPrinter:       trafficPolicyPrinter,
		AccessControlPolicyPrinter: accessControlPolicyPrinter,
		VirtualMeshPrinter:         virtualMeshPrinter,
		VirtualMeshCSRPrinter:      vmcsrPrinter,
		ResourcePrinter:            resourcePrinter,
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
