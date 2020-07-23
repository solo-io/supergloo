package cluster

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/cluster/register"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Interacting with remote Kubernetes clusters registered to Service Mesh Hub",
	}

	cmd.AddCommand(
		register.Command(ctx),
	)

	return cmd
}
