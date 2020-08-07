package mesh

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/mesh/internal/flags"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/mesh/restart"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &flags.Options{}
	cmd := &cobra.Command{
		Use:   "mesh",
		Short: "Operations on a specific mesh",
	}
	opts.AddToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		restart.Command(ctx, opts),
	)

	return cmd
}
