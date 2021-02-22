package register

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/cluster/register/oss"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Kubernetes cluster with Gloo Mesh",
	}
	oss.Command(ctx, globalFlags)
	return cmd
}
