package common

import (
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/status"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/describe/description"
	upgrade_assets "github.com/solo-io/service-mesh-hub/cli/pkg/tree/upgrade/assets"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/version/server"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	k8s_apiextensions "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apiextensions.k8s.io/v1beta1"
	k8s_apps_v1_clients "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/cluster-registration"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/version"
	"github.com/solo-io/service-mesh-hub/pkg/kube/auth"
	crd_uninstall "github.com/solo-io/service-mesh-hub/pkg/kube/crd"
	"github.com/solo-io/service-mesh-hub/pkg/kube/helm"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/kube/unstructured"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator"
	"k8s.io/client-go/rest"
)

// a grab bag of various clients that command implementations may use
type KubeClients struct {
	ClusterAuthorization        auth.ClusterAuthorization
	HelmInstallerFactory        helm.HelmInstallerFactory
	HelmClientFileConfigFactory helm.HelmClientForFileConfigFactory
	KubeClusterClient           smh_discovery.KubernetesClusterClient // client for KubernetesCluster custom resources
	MeshServiceClient           smh_discovery.MeshServiceClient
	MeshWorkloadClient          smh_discovery.MeshWorkloadClient
	MeshClient                  smh_discovery.MeshClient
	VirtualMeshClient           smh_networking.VirtualMeshClient
	VirtualMeshCSRClient        smh_security.VirtualMeshCertificateSigningRequestClient
	DeployedVersionFinder       version.DeployedVersionFinder
	CrdClientFactory            k8s_apiextensions.CustomResourceDefinitionClientFromConfigFactory
	HealthCheckClients          healthcheck_types.Clients
	SecretClient                k8s_core.SecretClient
	NamespaceClient             k8s_core.NamespaceClient
	UninstallClients            UninstallClients
	ClusterDeregistrationClient cluster_registration.ClusterDeregistrationClient
	KubeConfigLookup            kubeconfig.KubeConfigLookup
	TrafficPolicyClient         smh_networking.TrafficPolicyClient
	AccessControlPolicyClient   smh_networking.AccessControlPolicyClient
	ResourceDescriber           description.ResourceDescriber
	ResourceSelector            selection.ResourceSelector
	ClusterRegistrationClient   cluster_registration.ClusterRegistrationClient
	DeploymentClient            k8s_apps_v1_clients.DeploymentClient
}

type KubeClientsFactory func(masterConfig *rest.Config, writeNamespace string) (*KubeClients, error)

type Clients struct {
	ServerVersionClient           server.ServerVersionClient
	MasterClusterVerifier         common_config.MasterKubeConfigVerifier
	ReleaseAssetHelper            upgrade_assets.AssetHelper
	UnstructuredKubeClientFactory unstructured.UnstructuredKubeClientFactory
	DeploymentClient              server.DeploymentClient
	StatusClientFactory           status.StatusClientFactory
	HealthCheckSuite              healthcheck_types.HealthCheckSuite
	KubeConverter                 kubeconfig.Converter
	IstioClients                  IstioClients
}

func IstioClientsProvider(
	operatorManagerFactory operator.OperatorManagerFactory,
	operatorDaoFactory operator.OperatorDaoFactory,
	operatorManifestBuilder operator.InstallerManifestBuilder,
) IstioClients {
	return IstioClients{
		OperatorManagerFactory:  operatorManagerFactory,
		OperatorDaoFactory:      operatorDaoFactory,
		OperatorManifestBuilder: operatorManifestBuilder,
	}
}

type IstioClients struct {
	OperatorManagerFactory  operator.OperatorManagerFactory
	OperatorDaoFactory      operator.OperatorDaoFactory
	OperatorManifestBuilder operator.InstallerManifestBuilder
}

type UninstallClients struct {
	CrdRemover              crd_uninstall.CrdRemover
	SecretToConfigConverter kubeconfig.Converter
}

func UninstallClientsProvider(
	crdRemover crd_uninstall.CrdRemover,
	secretToConfigConverter kubeconfig.Converter,
) UninstallClients {
	return UninstallClients{
		CrdRemover:              crdRemover,
		SecretToConfigConverter: secretToConfigConverter,
	}
}

type ClientsFactory func(opts *options.Options) (*Clients, error)

func ClientsProvider(
	serverVersionClient server.ServerVersionClient,
	assetHelper upgrade_assets.AssetHelper,
	verifier common_config.MasterKubeConfigVerifier,
	unstructuredKubeClientFactory unstructured.UnstructuredKubeClientFactory,
	deploymentClient server.DeploymentClient,
	istioClients IstioClients,
	statusClientFactory status.StatusClientFactory,
	healthCheckSuite healthcheck_types.HealthCheckSuite,
	kubeConverter kubeconfig.Converter,
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
	helmInstallerFactory helm.HelmInstallerFactory,
	helmClientFileConfigFactory helm.HelmClientForFileConfigFactory,
	kubeClusterClient smh_discovery.KubernetesClusterClient,
	healthCheckClients healthcheck_types.Clients,
	deployedVersionFinder version.DeployedVersionFinder,
	crdClientFactory k8s_apiextensions.CustomResourceDefinitionClientFromConfigFactory,
	secretClient k8s_core.SecretClient,
	namespaceClient k8s_core.NamespaceClient,
	uninstallClients UninstallClients,
	clusterDeregistrationClient cluster_registration.ClusterDeregistrationClient,
	kubeConfigLookup kubeconfig.KubeConfigLookup,
	virtualMeshCsrClient smh_security.VirtualMeshCertificateSigningRequestClient,
	meshServiceClient smh_discovery.MeshServiceClient,
	meshClient smh_discovery.MeshClient,
	virtualMeshClient smh_networking.VirtualMeshClient,
	resourceDescriber description.ResourceDescriber,
	resourceSelector selection.ResourceSelector,
	trafficPolicyClient smh_networking.TrafficPolicyClient,
	accessControlPolicyClient smh_networking.AccessControlPolicyClient,
	meshWorkloadClient smh_discovery.MeshWorkloadClient,
	clusterRegistrationClient cluster_registration.ClusterRegistrationClient,
	deploymentClient k8s_apps_v1_clients.DeploymentClient,
) *KubeClients {
	return &KubeClients{
		ClusterAuthorization:        authorization,
		HelmInstallerFactory:        helmInstallerFactory,
		HelmClientFileConfigFactory: helmClientFileConfigFactory,
		KubeClusterClient:           kubeClusterClient,
		MeshServiceClient:           meshServiceClient,
		VirtualMeshCSRClient:        virtualMeshCsrClient,
		DeployedVersionFinder:       deployedVersionFinder,
		CrdClientFactory:            crdClientFactory,
		HealthCheckClients:          healthCheckClients,
		SecretClient:                secretClient,
		NamespaceClient:             namespaceClient,
		UninstallClients:            uninstallClients,
		ClusterDeregistrationClient: clusterDeregistrationClient,
		KubeConfigLookup:            kubeConfigLookup,
		MeshClient:                  meshClient,
		VirtualMeshClient:           virtualMeshClient,
		ResourceDescriber:           resourceDescriber,
		ResourceSelector:            resourceSelector,
		TrafficPolicyClient:         trafficPolicyClient,
		AccessControlPolicyClient:   accessControlPolicyClient,
		MeshWorkloadClient:          meshWorkloadClient,
		ClusterRegistrationClient:   clusterRegistrationClient,
		DeploymentClient:            deploymentClient,
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
