package demo

import (
	"context"

	istio_multicluster "github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo/istio-multicluster"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Command line utilities for running/interacting with Service Mesh Hub demos",
	}

	cmd.AddCommand(
		istio_multicluster.Command(ctx),
	)

	return cmd
}
