package istio_multicluster

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo/istio-multicluster/cleanup"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo/istio-multicluster/install"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "istio-multicluster",
		Short: "Demo Service Mesh Hub functionality with two Istio control planes deployed on separate clusters",
	}

	cmd.AddCommand(
		install.Command(ctx),
		cleanup.Command(ctx),
	)

	return cmd
}
