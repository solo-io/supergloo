package demo

import (
	"context"

	istio_multicluster "github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo/istio-multicluster"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo/osm"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Bootstrap environments for various demos demonstrating Service Mesh Hub functionality.",
	}

	cmd.AddCommand(
		istio_multicluster.Command(ctx),
		osm.Command(ctx),
	)

	return cmd
}
