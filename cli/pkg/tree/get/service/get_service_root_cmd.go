package get_service

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

type GetServiceCommand *cobra.Command

var (
	GetServiceSet = wire.NewSet(
		GetServiceRootCommand,
	)
)

func GetServiceRootCommand(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	kubeClientFactory common.KubeClientsFactory,
	kubeLoader common_config.KubeLoader,
	opts *options.Options,
) GetServiceCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.GetCommand.Service.Use,
		Aliases: cliconstants.GetCommand.Service.Aliases,
		Short:   cliconstants.GetCommand.Service.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetMeshServices(ctx, out, printers, kubeClientFactory, kubeLoader, opts)
		},
	}
	return cmd
}
