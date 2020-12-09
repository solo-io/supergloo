package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/dashboard"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/plugins"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/check"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/cluster"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/demo"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/describe"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/mesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/uninstall"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/version"

	"github.com/spf13/cobra"

	// required import to enable kube client-go auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const binaryName = "meshctl"

func RootCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meshctl [command]",
		Short: "The Command Line Interface for managing Gloo Mesh.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.DebugLevel)
		},
	}

	cmd.AddCommand(
		cluster.Command(ctx),
		demo.Command(ctx),
		debug.Command(ctx),
		describe.Command(ctx),
		mesh.Command(ctx),
		install.Command(ctx),
		uninstall.Command(ctx),
		check.Command(ctx),
		dashboard.Command(ctx),
		version.Command(ctx),
	)

	if len(os.Args) > 1 {
		if _, _, err := cmd.Find(os.Args[1:]); err != nil {
			handler := plugins.NewPathHandler(binaryName)
			if err := plugins.Handle(handler, os.Args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "plugin error: %s\n", err.Error())
				os.Exit(1)
			}

			os.Exit(0)
		}
	}

	return cmd
}
