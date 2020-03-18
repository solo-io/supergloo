// +build wireinject

package wire

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/go-utils/installutils/helminstall"
	cli "github.com/solo-io/mesh-projects/cli/pkg"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/common/kube"
	"github.com/solo-io/mesh-projects/cli/pkg/common/usage"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/check"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/check/healthcheck"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/check/status"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	register "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register/csr"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/install"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/auth"
	kubernetes_apps "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/apps"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	kubernetes_discovery "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/discovery"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	version2 "github.com/solo-io/mesh-projects/pkg/version"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func DefaultKubeClientsFactory(masterConfig *rest.Config, writeNamespace string) (clients *common.KubeClients, err error) {
	wire.Build(
		kubernetes.NewForConfig,
		wire.Bind(new(kubernetes.Interface), new(*kubernetes.Clientset)),
		kubernetes_core.NewGeneratedServiceAccountClient,
		discovery_core.NewGeneratedKubernetesClusterClient,
		kubernetes_core.NewGeneratedSecretsClient,
		kubernetes_core.NewGeneratedNamespaceClient,
		kubernetes_discovery.NewGeneratedServerVersionClient,
		kubernetes_core.NewGeneratedPodClient,
		discovery_core.NewGeneratedMeshServiceClient,
		kubernetes_apps.NewGeneratedDeploymentClient,
		auth.NewRemoteAuthorityConfigCreator,
		auth.RbacClientProvider,
		auth.NewRemoteAuthorityManager,
		common.DefaultSecretWriterProvider,
		auth.NewClusterAuthorization,
		docker.NewImageNameParser,
		version2.NewDeployedVersionFinder,
		helminstall.DefaultHelmClient,
		install.HelmInstallerProvider,
		healthcheck.ClientsProvider,
		common.KubeClientsProvider,
	)
	return nil, nil
}

func DefaultClientsFactory(opts *options.Options) (*common.Clients, error) {
	wire.Build(
		upgrade.UpgraderClientSet,
		docker.NewImageNameParser,
		common_config.DefaultKubeLoaderProvider,
		common_config.NewMasterKubeConfigVerifier,
		server.DefaultServerVersionClientProvider,
		kube.NewUnstructuredKubeClientFactory,
		server.NewDeploymentClient,
		operator.NewInstallerManifestBuilder,
		common.IstioClientsProvider,
		register.NewCsrAgentInstallerFactory,
		common.ClusterRegistrationClientsProvider,
		operator.NewOperatorManagerFactory,
		status.StatusClientFactoryProvider,
		healthcheck.DefaultHealthChecksProvider,
		common.ClientsProvider,
	)
	return nil, nil
}

func InitializeCLI(ctx context.Context, out io.Writer) *cobra.Command {
	wire.Build(
		docker.NewImageNameParser,
		common.NewDefaultFileReader,
		common_config.DefaultKubeLoaderProvider,
		options.NewOptionsProvider,
		DefaultKubeClientsFactoryProvider,
		DefaultClientsFactoryProvider,
		usage.DefaultUsageReporterProvider,
		upgrade.UpgradeSet,
		cluster.ClusterSet,
		version.VersionSet,
		istio.IstioProviderSet,
		install.InstallSet,
		check.CheckSet,
		cli.BuildCli,
	)
	return nil
}

func InitializeCLIWithMocks(
	ctx context.Context,
	out io.Writer,
	usageClient usageclient.Client,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	kubeLoader common_config.KubeLoader,
	imageNameParser docker.ImageNameParser,
	fileReader common.FileReader,
) *cobra.Command {

	wire.Build(
		options.NewOptionsProvider,
		cluster.ClusterSet,
		version.VersionSet,
		istio.IstioProviderSet,
		install.InstallSet,
		upgrade.UpgradeSet,
		check.CheckSet,
		cli.BuildCli,
	)
	return nil
}
