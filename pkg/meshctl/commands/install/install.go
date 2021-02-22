package install

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/oss"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oss",
		Short: "Install Gloo Mesh",
	}

	cmd.AddCommand(oss.Command(ctx, globalFlags))
	cmd.AddCommand(enterprise.Command(ctx, globalFlags))

	return cmd
}
