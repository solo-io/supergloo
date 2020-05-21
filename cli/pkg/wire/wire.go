// +build wireinject

package wire

import (
	"context"
	"io"

	"github.com/google/wire"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	cli "github.com/solo-io/service-mesh-hub/cli/pkg"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/files"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/usage"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/status"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/demo"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/describe"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/describe/description"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/get"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/install"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio/operator"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/config_lookup"
	crd_uninstall "github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/crd"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/upgrade"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/version"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/version/server"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	kubernetes_apiext "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apiextensions.k8s.io/v1beta1"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/auth"
	clients2 "github.com/solo-io/service-mesh-hub/pkg/clients"
	cluster_registration "github.com/solo-io/service-mesh-hub/pkg/clients/cluster-registration"
	kubernetes_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/discovery"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	"github.com/solo-io/service-mesh-hub/pkg/factories"
	"github.com/solo-io/service-mesh-hub/pkg/installers/csr"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	version2 "github.com/solo-io/service-mesh-hub/pkg/version"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func DefaultKubeClientsFactory(masterConfig *rest.Config, writeNamespace string) (clients *common.KubeClients, err error) {
	wire.Build(
		kubernetes.NewForConfig,
		wire.Bind(new(kubernetes.Interface), new(*kubernetes.Clientset)),
		kubernetes_discovery.NewGeneratedServerVersionClient,
		k8s_core.ClientsetFromConfigProvider,
		k8s_core.ServiceAccountClientFromClientsetProvider,
		k8s_core.SecretClientFromClientsetProvider,
		k8s_core.NamespaceClientFromClientsetProvider,
		k8s_core.PodClientFromClientsetProvider,
		k8s_core.SecretClientFactoryProvider,
		k8s_core.ServiceAccountClientFactoryProvider,
		k8s_core.NamespaceClientFromConfigFactoryProvider,
		kubernetes_apps.ClientsetFromConfigProvider,
		kubernetes_apps.DeploymentClientFromClientsetProvider,
		kubernetes_apps.DeploymentClientFactoryProvider,
		kubernetes_apiext.CustomResourceDefinitionClientFromConfigFactoryProvider,
		files.NewDefaultFileReader,
		auth.NewRemoteAuthorityConfigCreator,
		auth.RbacClientProvider,
		auth.NewRemoteAuthorityManager,
		auth.NewClusterAuthorization,
		docker.NewImageNameParser,
		version2.NewDeployedVersionFinder,
		install.HelmInstallerProvider,
		healthcheck.ClientsProvider,
		crd_uninstall.NewCrdRemover,
		kubeconfig.NewConverter,
		common.UninstallClientsProvider,
		kubeconfig.NewKubeConfigLookup,
		config_lookup.NewDynamicClientGetter,
		cluster_registration.NewClusterDeregistrationClient,
		common.KubeClientsProvider,
		description.NewResourceDescriber,
		selector.NewResourceSelector,
		zephyr_discovery.ClientsetFromConfigProvider,
		zephyr_networking.ClientsetFromConfigProvider,
		zephyr_security.ClientsetFromConfigProvider,
		zephyr_discovery.KubernetesClusterClientFromClientsetProvider,
		zephyr_discovery.MeshServiceClientFromClientsetProvider,
		zephyr_discovery.MeshWorkloadClientFromClientsetProvider,
		zephyr_discovery.MeshClientFromClientsetProvider,
		zephyr_networking.TrafficPolicyClientFromClientsetProvider,
		zephyr_networking.AccessControlPolicyClientFromClientsetProvider,
		zephyr_networking.VirtualMeshClientFromClientsetProvider,
		zephyr_security.VirtualMeshCertificateSigningRequestClientFromClientsetProvider,
		csr.NewCsrAgentInstallerFactory,
		factories.HelmClientForMemoryConfigFactoryProvider,
		factories.HelmClientForFileConfigFactoryProvider,
		cluster_registration.NewClusterRegistrationClient,
		clients2.ClusterAuthClientFromConfigFactoryProvider,
	)
	return nil, nil
}

func DefaultClientsFactory(opts *options.Options) (*common.Clients, error) {
	wire.Build(
		files.NewDefaultFileReader,
		kubeconfig.NewConverter,
		upgrade.UpgraderClientSet,
		docker.NewImageNameParser,
		kubeconfig.DefaultKubeLoaderProvider,
		common_config.NewMasterKubeConfigVerifier,
		server.DefaultServerVersionClientProvider,
		kube.NewUnstructuredKubeClientFactory,
		server.NewDeploymentClient,
		operator.NewInstallerManifestBuilder,
		common.IstioClientsProvider,
		operator.NewOperatorManagerFactory,
		status.StatusClientFactoryProvider,
		healthcheck.DefaultHealthChecksProvider,
		common.ClientsProvider,
	)
	return nil, nil
}

func InitializeCLI(ctx context.Context, out io.Writer, in io.Reader) *cobra.Command {
	wire.Build(
		docker.NewImageNameParser,
		files.NewDefaultFileReader,
		kubeconfig.DefaultKubeLoaderProvider,
		options.NewOptionsProvider,
		DefaultKubeClientsFactoryProvider,
		DefaultClientsFactoryProvider,
		table_printing.TablePrintingSet,
		resource_printing.NewResourcePrinter,
		common.PrintersProvider,
		usage.DefaultUsageReporterProvider,
		interactive.NewSurveyInteractivePrompt,
		exec.NewShellRunner,
		demo.DemoSet,
		upgrade.UpgradeSet,
		cluster.ClusterSet,
		version.VersionSet,
		mesh.MeshProviderSet,
		install.InstallSet,
		uninstall.UninstallSet,
		check.CheckSet,
		describe.DescribeSet,
		create.CreateSet,
		get.GetSet,
		cli.BuildCli,
		afero.NewOsFs,
	)
	return nil
}

func InitializeCLIWithMocks(
	ctx context.Context,
	out io.Writer,
	in io.Reader,
	usageClient usageclient.Client,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	kubeLoader kubeconfig.KubeLoader,
	imageNameParser docker.ImageNameParser,
	fileReader files.FileReader,
	kubeconfigConverter kubeconfig.Converter,
	printers common.Printers,
	runner exec.Runner,
	interactivePrompt interactive.InteractivePrompt,
	fs afero.Fs,
) *cobra.Command {
	wire.Build(
		options.NewOptionsProvider,
		demo.DemoSet,
		cluster.ClusterSet,
		version.VersionSet,
		mesh.MeshProviderSet,
		install.InstallSet,
		upgrade.UpgradeSet,
		uninstall.UninstallSet,
		check.CheckSet,
		get.GetSet,
		describe.DescribeSet,
		create.CreateSet,
		cli.BuildCli,
	)
	return nil
}
