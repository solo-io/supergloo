package create

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/common/resource_printing"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/create/virtualmesh"
	"github.com/spf13/cobra"
)

var (
	CreateSet = wire.NewSet(
		CreateRootCommand,
		virtualmesh.CreateVirtualMeshCommand,
	)
)

type CreateRootCmd *cobra.Command

func CreateRootCommand(
	opts *options.Options,
	createVirtualMeshCmd virtualmesh.CreateVirtualMeshCmd,
) CreateRootCmd {
	cmd := &cobra.Command{
		Use:   cliconstants.CreateCommand.Use,
		Short: cliconstants.CreateCommand.Short,
		RunE:  common.NonTerminalCommand(cliconstants.CreateCommand.Use),
	}
	options.AddCreateFlags(cmd, opts, resource_printing.YAMLFormat.String(), resource_printing.ValidFormats)
	cmd.AddCommand(createVirtualMeshCmd)
	return cmd
}
