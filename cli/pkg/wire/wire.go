// +build wireinject

package wire

import (
	"context"
	"io"

	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"

	"github.com/google/wire"
	cli "github.com/solo-io/mesh-projects/cli/pkg"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/pkg/auth"
	"github.com/spf13/cobra"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func DefaultKubeClientsFactory(masterConfig *rest.Config, writeNamespace string) (*common.KubeClients, error) {
	wire.Build(
		auth.DefaultClientsProvider,
		auth.NewRemoteAuthorityConfigCreator,
		auth.NewRemoteAuthorityManager,
		k8sclientv1.NewForConfig,
		common.DefaultSecretWriterProvider,
		auth.NewClusterAuthorization,
		common.KubeClientsProvider,
	)
	return nil, nil
}

func DefaultClientsFactory(opts *options.Options) (*common.Clients, error) {
	wire.Build(
		common_config.DefaultKubeLoaderProvider,
		common_config.DefaultFileExistenceCheckerProvider,
		common_config.NewMasterKubeConfigVerifier,
		server.DefaultServerVersionClientProvider,
		common.ClientsProvider,
	)
	return nil, nil
}

func InitializeCLI(ctx context.Context, out io.Writer) *cobra.Command {
	wire.Build(
		options.NewOptionsProvider,
		DefaultKubeClientsFactoryProvider,
		DefaultClientsFactoryProvider,
		common.DefaultUsageReporterProvider,
		cluster.ClusterSet,
		version.VersionSet,
		cli.BuildCli,
	)
	return nil
}
