package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/check"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/cluster"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/dashboard"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/debug"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/demo"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/describe"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/initpluginmanager"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/mesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/uninstall"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/version"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/plugins"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"

	"github.com/spf13/cobra"

	// required import to enable kube client-go auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const binaryName = "meshctl"

func RootCommand(ctx context.Context) *cobra.Command {
	globalFlags := &utils.GlobalFlags{}

	cmd := &cobra.Command{
		Use:   "meshctl [command]",
		Short: "The Command Line Interface for managing Gloo Mesh.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if globalFlags.Verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}
		},
	}

	// set global CLI flags
	globalFlags.AddToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		cluster.Command(ctx, globalFlags),
		demo.Command(ctx),
		debug.Command(ctx, globalFlags),
		describe.Command(ctx),
		mesh.Command(ctx),
		install.Command(ctx, globalFlags),
		uninstall.Command(ctx, globalFlags),
		check.Command(ctx),
		dashboard.Command(ctx),
		version.Command(ctx),
		initpluginmanager.Command(ctx),
	)

	if len(os.Args) > 1 {
		if _, _, err := cmd.Find(os.Args[1:]); err != nil {
			handler := plugins.NewPathHandler(binaryName)
			if err := plugins.Handle(handler, os.Args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "plugin error: %s\n", err.Error())
				os.Exit(1)
			}
		}
	}

	return cmd
}
