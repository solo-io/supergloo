package describe

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/accesspolicy"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/failoverservice"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/mesh"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/trafficpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/traffictarget"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/virtualmesh"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/workload"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Human readable description of discovered resources and applicable configuration",
	}

	cmd.AddCommand(
		mesh.Command(ctx),
		traffictarget.Command(ctx),
		failoverservice.Command(ctx),
		virtualmesh.Command(ctx),
		workload.Command(ctx),
		accesspolicy.Command(ctx),
		trafficpolicy.Command(ctx),
	)

	return cmd
}
