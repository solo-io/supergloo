package get_mesh

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/spf13/cobra"
)

type GetMeshCommand *cobra.Command

var (
	GetMeshSet = wire.NewSet(
		GetMeshRootCommand,
	)
)

func GetMeshRootCommand(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	kubeClientFactory common.KubeClientsFactory,
	kubeLoader common_config.KubeLoader,
	opts *options.Options,
) GetMeshCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.GetCommand.Mesh.Use,
		Aliases: cliconstants.GetCommand.Mesh.Aliases,
		Short:   cliconstants.GetCommand.Mesh.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetMeshes(ctx, out, printers, kubeClientFactory, kubeLoader, opts)
		},
	}
	return cmd
}
