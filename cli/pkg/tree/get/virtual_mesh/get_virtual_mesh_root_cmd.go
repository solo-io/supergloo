package get_virtual_mesh

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

type GetVirtualMeshCommand *cobra.Command

var (
	GetVirtualMeshSet = wire.NewSet(
		GetVirtualMeshRootCommand,
	)
)

func GetVirtualMeshRootCommand(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	kubeClientFactory common.KubeClientsFactory,
	kubeLoader common_config.KubeLoader,
	opts *options.Options,
) GetVirtualMeshCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.GetCommand.VirtualMesh.Use,
		Aliases: cliconstants.GetCommand.VirtualMesh.Aliases,
		Short:   cliconstants.GetCommand.VirtualMesh.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetVirtualMeshes(ctx, out, printers, kubeClientFactory, kubeLoader, opts)
		},
	}
	return cmd
}
