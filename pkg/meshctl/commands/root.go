package commands

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/cluster"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/mesh"
	"github.com/spf13/cobra"
)

func RootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "meshctl [command]",
		Short:   "The Command Line Interface for managing Service Mesh Hub.",
		Version: version.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.DebugLevel)
		},
	}

	cmd.AddCommand(
		cluster.Command(ctx),
		mesh.Command(ctx),
	)

	return cmd
}
