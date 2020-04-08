package get_workload

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

type GetWorkloadCommand *cobra.Command

var (
	GetWorkloadSet = wire.NewSet(
		GetWorkloadRootCommand,
	)
)

func GetWorkloadRootCommand(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	kubeClientFactory common.KubeClientsFactory,
	kubeLoader common_config.KubeLoader,
	opts *options.Options,
) GetWorkloadCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.GetCommand.Workload.Use,
		Aliases: cliconstants.GetCommand.Workload.Aliases,
		Short:   cliconstants.GetCommand.Workload.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetMeshWorkloads(ctx, out, printers, kubeClientFactory, kubeLoader, opts)
		},
	}
	return cmd
}
