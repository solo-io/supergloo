package commands

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/check"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/cluster"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/demo"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/install"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/mesh"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/uninstall"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/version"

	"github.com/spf13/cobra"
)

func RootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meshctl [command]",
		Short: "The Command Line Interface for managing Service Mesh Hub.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.DebugLevel)
		},
	}

	cmd.AddCommand(
		cluster.Command(ctx),
		demo.Command(ctx),
		describe.Command(ctx),
		mesh.Command(ctx),
		install.Command(ctx),
		uninstall.Command(ctx),
		check.Command(ctx),
		version.Command(ctx),
	)

	return cmd
}
