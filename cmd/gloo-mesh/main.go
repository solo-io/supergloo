package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	mesh_discovery "github.com/solo-io/gloo-mesh/pkg/mesh-discovery"
	mesh_networking "github.com/solo-io/gloo-mesh/pkg/mesh-networking"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()

	if err := rootCommand(ctx).Execute(); err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
	contextutils.LoggerFrom(ctx).Info("exiting...")
}

func rootCommand(ctx context.Context) *cobra.Command {
	opts := &bootstrap.Options{}
	cmd := &cobra.Command{
		Use:     "gloo-mesh [command]",
		Short:   "Start the Gloo Mesh Operators (discovery and networking)",
		Version: version.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.DebugLevel)
		},
	}

	opts.AddToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		discoveryCommand(ctx, opts),
		networkingCommand(ctx, opts),
	)

	return cmd
}

type discoveryOpts struct {
	*bootstrap.Options
}

func discoveryCommand(ctx context.Context, bs *bootstrap.Options) *cobra.Command {
	opts := &discoveryOpts{Options: bs}
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Start the Gloo Mesh Discovery Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startDiscovery(ctx, opts)
		},
	}
	return cmd
}

func startDiscovery(ctx context.Context, opts *discoveryOpts) error {
	return mesh_discovery.Start(ctx, *opts.Options)
}

type networkingOpts struct {
	*bootstrap.Options
}

func networkingCommand(ctx context.Context, bs *bootstrap.Options) *cobra.Command {
	opts := &networkingOpts{Options: bs}
	cmd := &cobra.Command{
		Use:   "networking",
		Short: "Start the Gloo Mesh Networking Operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startNetworking(ctx, opts)
		},
	}
	return cmd
}

func startNetworking(ctx context.Context, opts *networkingOpts) error {
	return mesh_networking.Start(ctx, *opts.Options)
}
