package get_cluster

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
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
	kubeLoader kubeconfig.KubeLoader,
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
