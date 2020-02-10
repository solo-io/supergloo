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
	"github.com/solo-io/mesh-projects/cli/pkg/tree/install"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	"github.com/spf13/cobra"
)

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
		cli.BuildCli,
	)
	return nil
}
