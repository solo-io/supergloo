package common

import (
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
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	"github.com/solo-io/service-mesh-hub/pkg/version"
	"k8s.io/client-go/rest"
)

// a grab bag of various clients that command implementations may use
type KubeClients struct {
	ClusterAuthorization            auth.ClusterAuthorization
	HelmInstallerFactory            factories.HelmerInstallerFactory
	HelmClientFileConfigFactory     factories.HelmClientForFileConfigFactory
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
	ClusterRegistrationClients      ClusterRegistrationClients
	ClusterRegistrationClient       clients.ClusterRegistrationClient
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
	KubeConverter                 kube.Converter
	IstioClients                  IstioClients
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
	SecretToConfigConverter kube.Converter
}

func UninstallClientsProvider(
	crdRemover crd_uninstall.CrdRemover,
	secretToConfigConverter kube.Converter,
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
	kubeConverter kube.Converter,
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
		KubeConverter:                 kubeConverter,
	}
}

// facilitates wire codegen
func KubeClientsProvider(
	authorization auth.ClusterAuthorization,
	helmInstallerFactory factories.HelmerInstallerFactory,
	helmClientFileConfigFactory factories.HelmClientForFileConfigFactory,
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
	clusterRegistrationClients ClusterRegistrationClients,
	clusterRegistrationClient clients.ClusterRegistrationClient,
) *KubeClients {
	return &KubeClients{
		ClusterAuthorization:            authorization,
		HelmInstallerFactory:            helmInstallerFactory,
		HelmClientFileConfigFactory:     helmClientFileConfigFactory,
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
		ClusterRegistrationClients:      clusterRegistrationClients,
		ClusterRegistrationClient:       clusterRegistrationClient,
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
