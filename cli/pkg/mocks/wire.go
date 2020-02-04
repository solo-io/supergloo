// +build wireinject

package cli_mocks

import (
	"context"
	"io"

	"github.com/google/wire"
	cli "github.com/solo-io/mesh-projects/cli/pkg"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade"
	upgrade_assets "github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade/assets"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	"github.com/solo-io/mesh-projects/pkg/auth"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	"github.com/spf13/cobra"
)

func InitializeCLIWithMocks(
	ctx context.Context,
	out io.Writer,
	usageClient usageclient.Client,
	authorization auth.ClusterAuthorization,
	writer common.SecretWriter,
	client server.ServerVersionClient,
	kubeLoader common_config.KubeLoader,
	verifier common_config.MasterKubeConfigVerifier,
	upgrader upgrade_assets.AssetHelper) *cobra.Command {
	wire.Build(
		options.NewOptionsProvider,
		MockKubeClientsFactoryProvider,
		MockClientsFactoryProvider,
		cluster.ClusterSet,
		version.VersionSet,
		upgrade.UpgradeSet,
		cli.BuildCli,
	)
	return nil
}
