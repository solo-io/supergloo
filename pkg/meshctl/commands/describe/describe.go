package describe

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/mesh"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/meshservice"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Human readable description of discovered resources and applicable configuration",
		RunE:  utils.NonTerminalCommand("describe"),
	}

	cmd.AddCommand(
		mesh.Command(ctx),
		meshservice.Command(ctx),
	)

	return cmd
}
