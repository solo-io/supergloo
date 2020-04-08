package demo

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/demo/cleanup"
	demo_init "github.com/solo-io/service-mesh-hub/cli/pkg/tree/demo/init"
	"github.com/spf13/cobra"
)

type DemoCommand *cobra.Command

var DemoSet = wire.NewSet(
	demo_init.InitSet,
	cleanup.CleanupSet,
	DemoRootCmd,
)

func DemoRootCmd(initCmd demo_init.InitCmd, cleanupCmd cleanup.CleanupCmd) DemoCommand {
	demo := &cobra.Command{
		Use:   cliconstants.DemoCommand.Use,
		Short: cliconstants.DemoCommand.Short,
		RunE:  common.NonTerminalCommand(cliconstants.DemoCommand.Use),
	}
	demo.AddCommand(
		cleanupCmd,
		initCmd,
	)
	return demo
}
