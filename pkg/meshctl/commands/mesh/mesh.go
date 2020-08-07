package mesh

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/mesh/flags"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/mesh/restart"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &flags.Options{}
	cmd := &cobra.Command{
		Use:   "mesh",
		Short: "Operations on a specific mesh",
		RunE:  utils.NonTerminalCommand("mesh"),
	}
	opts.AddToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		restart.Command(ctx, opts),
	)

	return cmd
}
