package describe

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/internal/flags"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/mesh"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/traffictarget"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &flags.Options{}
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Human readable description of discovered resources and applicable configuration",
	}
	opts.AddToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		mesh.Command(ctx, opts),
		traffictarget.Command(ctx, opts),
	)

	return cmd
}
