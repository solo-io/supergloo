package demo_init

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
	"github.com/spf13/cobra"
)

type InitCmd *cobra.Command

var InitSet = wire.NewSet(
	DemoInitCmd,
)

func DemoInitCmd(
	ctx context.Context,
	runnner exec.Runner,
) InitCmd {
	init := &cobra.Command{
		Use:   cliconstants.DemoInitCommand.Use,
		Short: cliconstants.DemoInitCommand.Short,
		Long:  cliconstants.DemoInitCommand.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return DemoInit(ctx, runnner)
		},
	}

	// options.AddClusterRegisterFlags(init, opts)
	return init
}
