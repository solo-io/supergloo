package debug

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/debug/snapshot"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug Service Mesh Hub resources",
	}

	cmd.AddCommand(
		snapshot.Command(ctx),
	)

	return cmd
}
