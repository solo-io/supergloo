package get_cluster

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/spf13/cobra"
)

type GetClusterCommand *cobra.Command

var (
	GetClusterSet = wire.NewSet(
		GetClusterRootCommand,
	)
)

func GetClusterRootCommand(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	kubeClientFactory common.KubeClientsFactory,
	kubeLoader common_config.KubeLoader,
	opts *options.Options,
) GetClusterCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.GetCommand.Cluster.Use,
		Aliases: cliconstants.GetCommand.Cluster.Aliases,
		Short:   cliconstants.GetCommand.Cluster.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetClusters(ctx, out, printers, kubeClientFactory, kubeLoader, opts)
		},
	}
	return cmd
}
