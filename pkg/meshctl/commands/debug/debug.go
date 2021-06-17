package debug

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/snapshot"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug Gloo Mesh resources",
	}

	cmd.AddCommand(
		snapshot.Command(ctx, globalFlags),
	)

	return cmd
}
