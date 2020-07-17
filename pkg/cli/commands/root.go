package commands

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/smh/pkg/cli/commands/cluster"
	"github.com/spf13/cobra"
)

func RootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "smh [command]",
		Short:   "The Command Line Tool for interacting with Service Mesh Hub",
		Version: version.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.DebugLevel)
		},
	}

	cmd.AddCommand(
		cluster.Command(ctx),
	)

	return cmd
}
