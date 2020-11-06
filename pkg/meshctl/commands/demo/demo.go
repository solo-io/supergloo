package demo

import (
	"context"

	istio_multicluster "github.com/solo-io/gloo-mesh/pkg/meshctl/commands/demo/istio-multicluster"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/demo/osm"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Bootstrap environments for various demos demonstrating Gloo Mesh functionality.",
	}

	cmd.AddCommand(
		istio_multicluster.Command(ctx),
		osm.Command(ctx),
	)

	return cmd
}
