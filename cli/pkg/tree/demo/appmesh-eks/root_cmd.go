package appmesh_eks

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/spf13/cobra"
)

type AppmeshEksCmd *cobra.Command

var AppmeshEksSet = wire.NewSet(
	AppmeshEks,
	Init,
	Cleanup,
)

func AppmeshEks(
	initCmd InitCmd,
	cleanupCmd CleanupCmd,
) AppmeshEksCmd {
	init := &cobra.Command{
		Use:   cliconstants.AppmeshEksCommand.Use,
		Short: cliconstants.AppmeshEksCommand.Short,
		Long:  cliconstants.AppmeshEksCommand.Long,
		RunE:  common.NonTerminalCommand(cliconstants.GetCommand.Root.Use),
	}
	init.AddCommand(
		initCmd,
		cleanupCmd,
	)
	return init
}
