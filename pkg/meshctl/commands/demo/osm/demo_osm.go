package osm

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo/common/cleanup"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo/common/initialize"
	"github.com/spf13/cobra"
)

const (
	mgmtCluster = "mgmt-cluster"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "osm",
		Short: "Demo Service Mesh Hub functionality one OSM control plane deployed",
		Long: `
Demo Service Mesh Hub functionality with one OSM control plane deployed on a local KinD cluster.

Requires kubectl >= v1.18.8, kind >= v0.8.1, osm >= v0.3.0, and docker.
We recommend allocating at least 8GB of RAM for Docker.
`,
	}

	cmd.AddCommand(
		initialize.OsmCommand(ctx, mgmtCluster),
		cleanup.Command(ctx, mgmtCluster),
	)

	return cmd
}
