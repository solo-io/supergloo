package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	mesh_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery"
	mesh_networking "github.com/solo-io/service-mesh-hub/pkg/mesh-networking"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	ctx := context.Background()

	if err := rootCommand(ctx).Execute(); err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
	contextutils.LoggerFrom(ctx).Info("exiting...")
}

type bootstrapOpts bootstrap.Options

func (opts bootstrapOpts) getBootstrap() bootstrap.Options {
	return bootstrap.Options(opts)
}

func (opts *bootstrapOpts) addToFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&opts.MasterNamespace, "namespace", "n", metav1.NamespaceAll, "if specified restricts the master manager's cache to watch objects in the desired namespace.")
	flags.Uint32Var(&opts.MetricsBindPort, "metrics-port", defaults.MetricsPort, "port on which to serve Prometheus metrics. set to 0 to disable")
	flags.BoolVar(&opts.VerboseMode, "verbose", true, "enables verbose/debug logging")
	flags.StringVar(&opts.SettingsName, "settings-name", defaults.DefaultSettingsName, "The name of the Settings object this controller should use.")
	flags.StringVar(&opts.SettingsNamespace, "settings-namespace", defaults.DefaultPodNamespace, "The namespace of the Settings object this controller should use.")
}

func rootCommand(ctx context.Context) *cobra.Command {
	opts := &bootstrapOpts{}
	cmd := &cobra.Command{
		Use:     "service-mesh-hub [command]",
		Short:   "Start the Service Mesh Hub Operators (discovery and networking)",
		Version: version.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.DebugLevel)
		},
	}

	opts.addToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		discoveryCommand(ctx, opts),
		networkingCommand(ctx, opts),
	)

	return cmd
}

type discoveryOpts struct {
	*bootstrapOpts
}

func discoveryCommand(ctx context.Context, bs *bootstrapOpts) *cobra.Command {
	opts := &discoveryOpts{bootstrapOpts: bs}
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Start the Service Mesh Hub Discovery Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startDiscovery(ctx, opts)
		},
	}
	return cmd
}

func startDiscovery(ctx context.Context, opts *discoveryOpts) error {
	return mesh_discovery.Start(ctx, opts.getBootstrap())
}

type networkingOpts struct {
	*bootstrapOpts
}

func networkingCommand(ctx context.Context, bs *bootstrapOpts) *cobra.Command {
	opts := &networkingOpts{bootstrapOpts: bs}
	cmd := &cobra.Command{
		Use:   "networking",
		Short: "Start the Service Mesh Hub Networking Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startNetworking(ctx, opts)
		},
	}
	return cmd
}

func startNetworking(ctx context.Context, opts *networkingOpts) error {
	return mesh_networking.Start(ctx, opts.getBootstrap())
}
