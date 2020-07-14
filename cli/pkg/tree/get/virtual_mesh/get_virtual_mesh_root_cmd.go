package get_virtual_mesh

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
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
	kubeLoader kubeconfig.KubeLoader,
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
