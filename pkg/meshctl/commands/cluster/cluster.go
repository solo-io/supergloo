package cluster

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/cluster/configure"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/cluster/deregister"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/cluster/list"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/cluster/register"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kubernetes clusters registered to Gloo Mesh",
	}

	cmd.AddCommand(
		register.Command(ctx, globalFlags),
		deregister.Command(ctx, globalFlags),
		list.Command(ctx),
		configure.Command(ctx),
	)

	return cmd
}
