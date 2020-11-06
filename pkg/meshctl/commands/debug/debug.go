package debug

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug/snapshot"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug Gloo Mesh resources",
	}

	cmd.AddCommand(
		snapshot.Command(ctx),
	)

	return cmd
}
