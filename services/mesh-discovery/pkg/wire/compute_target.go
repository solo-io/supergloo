package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register/csr"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	k8s_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"github.com/solo-io/service-mesh-hub/pkg/version"
	compute_target "github.com/solo-io/service-mesh-hub/services/common/compute-target"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	compute_target_aws "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	appmesh_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/clients/appmesh"
	eks_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/clients/eks"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s-cluster/rest/eks"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/appmesh"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/aws"
	"k8s.io/client-go/rest"
)

var AwsSet = wire.NewSet(
	compute_target_aws.NewAwsAPIHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	aws_utils.NewArnParser,
	aws_utils.NewAppMeshParser,
	appmesh.AppMeshWorkloadScannerFactoryProvider,
	zephyr_discovery.KubernetesClusterClientProvider,
	eks_client.EksClientFactoryProvider,
	eks_client.EksConfigBuilderFactoryProvider,
	appmesh_client.AppMeshClientFactoryProvider,
	AwsDiscoveryReconcilersProvider,
	aws.NewAppMeshDiscoveryReconciler,
	eks.NewEksDiscoveryReconciler,
)

var ClusterRegistrationSet = wire.NewSet(
	factories.HelmClientForMemoryConfigFactoryProvider,
	factories.HelmClientForFileConfigFactoryProvider,
	k8s_core.SecretClientFromConfigFactoryProvider,
	k8s_core.NamespaceClientFromConfigFactoryProvider,
	zephyr_discovery.KubernetesClusterClientFromConfigFactoryProvider,
	k8s_apps.DeploymentClientFromConfigFactoryProvider,
	k8s_core.ServiceAccountClientFromConfigFactoryProvider,
	auth.RbacClientFactoryProvider,
	auth.ClusterAuthorizationFactoryProvider,
	csr.NewCsrAgentInstallerFactory,
	DeployedVersionFinderProvider,
)

func AwsDiscoveryReconcilersProvider(
	appMeshReconciler compute_target_aws.AppMeshDiscoveryReconciler,
	eksReconciler compute_target_aws.EksDiscoveryReconciler,
) []compute_target_aws.RestAPIDiscoveryReconciler {
	return []compute_target_aws.RestAPIDiscoveryReconciler{appMeshReconciler, eksReconciler}
}

func ComputeTargetCredentialsHandlersProvider(
	asyncManagerController *mc_manager.AsyncManagerController,
	awsCredsHandler compute_target_aws.AwsCredsHandler,
) []compute_target.ComputeTargetCredentialsHandler {
	return []compute_target.ComputeTargetCredentialsHandler{
		asyncManagerController,
		awsCredsHandler,
	}
}

func DeployedVersionFinderProvider(
	masterCfg *rest.Config,
	deploymentClientFromConfigFactory k8s_apps.DeploymentClientFromConfigFactory,
	imageNameParser docker.ImageNameParser,
) (version.DeployedVersionFinder, error) {
	deploymentClient, err := deploymentClientFromConfigFactory(masterCfg)
	if err != nil {
		return nil, err
	}
	return version.NewDeployedVersionFinder(deploymentClient, imageNameParser), nil
}

func ClusterRegistrationClientProvider(
	masterCfg *rest.Config,
	secretClientFactory k8s_core.SecretClientFromConfigFactory,
	kubeClusterClient zephyr_discovery.KubernetesClusterClientFromConfigFactory,
	namespaceClientFactory k8s_core.NamespaceClientFromConfigFactory,
	kubeConverter kube.Converter,
	csrAgentInstallerFactory csr.CsrAgentInstallerFactory,
) (clients.ClusterRegistrationClient, error) {
	masterSecretClient, err := secretClientFactory(masterCfg)
	if err != nil {
		return nil, err
	}
	masterKubeClusterClient, err := kubeClusterClient(masterCfg)
	if err != nil {
		return nil, err
	}
	return clients.NewClusterRegistrationClient(
		masterSecretClient,
		masterKubeClusterClient,
		namespaceClientFactory,
		kubeConverter,
		csrAgentInstallerFactory,
		clients.DefaultClusterAuthClientFromConfig,
	), nil
}
