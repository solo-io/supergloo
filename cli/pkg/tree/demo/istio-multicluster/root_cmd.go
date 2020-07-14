package istio_multicluster

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/spf13/cobra"
)

type IstioMulticlusterCmd *cobra.Command

var IstioMulticlusterSet = wire.NewSet(
	IstioMulticluster,
	Init,
	Cleanup,
)

func IstioMulticluster(
	initCmd InitCmd,
	cleanupCmd CleanupCmd,
) IstioMulticlusterCmd {
	init := &cobra.Command{
		Use:   cliconstants.IstioMulticlusterCommand.Use,
		Short: cliconstants.IstioMulticlusterCommand.Short,
		Long:  cliconstants.IstioMulticlusterCommand.Long,
		RunE:  common.NonTerminalCommand(cliconstants.GetCommand.Root.Use),
	}
	init.AddCommand(
		initCmd,
		cleanupCmd,
	)
	return init
}
