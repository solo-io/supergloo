package commands

import (
	"bytes"
	"context"
	"fmt"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/bug-report/pkg/bugreport"
	"istio.io/pkg/log"
	"os"

	"github.com/fatih/color"
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
		SilenceErrors: true,
	}

	// Use custom logrus formatter
	logrus.SetFormatter(logFormatter{})

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
		bugreport.Cmd(log.DefaultOptions()),
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

type logFormatter struct{}

func (logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var buf bytes.Buffer
	switch entry.Level {
	case logrus.DebugLevel:
		color.New(color.FgRed).Fprintln(&buf, entry.Message)
	case logrus.InfoLevel:
		fmt.Fprintln(&buf, entry.Message)
	case logrus.WarnLevel:
		color.New(color.FgYellow).Fprint(&buf, "warning: ")
		fmt.Fprintln(&buf, entry.Message)
	case logrus.ErrorLevel:
		color.New(color.FgRed).Fprint(&buf, "error: ")
		fmt.Fprintln(&buf, entry.Message)
	case logrus.FatalLevel:
		color.New(color.FgRed).Fprint(&buf, "fatal: ")
		fmt.Fprintln(&buf, entry.Message)
	case logrus.PanicLevel:
		color.New(color.FgRed).Fprint(&buf, "panic: ")
		fmt.Fprintln(&buf, entry.Message)
	}

	return buf.Bytes(), nil
}
