package get_vmcsr

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

type GetVirtualMeshCSRCommand *cobra.Command

var (
	GetVirtualMeshCSRSet = wire.NewSet(
		GetVirtualMeshCSRRootCommand,
	)
)

func GetVirtualMeshCSRRootCommand(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	kubeClientFactory common.KubeClientsFactory,
	kubeLoader kubeconfig.KubeLoader,
	opts *options.Options,
) GetVirtualMeshCSRCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.GetCommand.VMCSR.Use,
		Aliases: cliconstants.GetCommand.VMCSR.Aliases,
		Short:   cliconstants.GetCommand.VMCSR.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetVirtualMeshCertificateSigningRequests(ctx, out, printers, kubeClientFactory, kubeLoader, opts)
		},
	}
	return cmd
}
