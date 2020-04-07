package cleanup

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
	"github.com/spf13/cobra"
)

type CleanupCmd *cobra.Command

var CleanupSet = wire.NewSet(
	DemoCleanupCmd,
)

func DemoCleanupCmd(
	ctx context.Context,
	runner exec.Runner,
) CleanupCmd {
	init := &cobra.Command{
		Use:   cliconstants.DemoCleanupCommand.Use,
		Short: cliconstants.DemoCleanupCommand.Short,
		Long:  cliconstants.DemoCleanupCommand.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return DemoCleanup(ctx, runner)
		},
	}

	// options.AddClusterRegisterFlags(init, opts)
	return init
}
