package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/agent"
	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
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
	flags.Uint32Var(&opts.MetricsBindPort, "metrics-port", 9091, "port on which to serve Prometheus metrics. set to 0 to disable")
	flags.BoolVar(&opts.VerboseMode, "verbose", true, "enables verbose/debug logging")
	flags.StringVarP(&opts.ManagementContext, "context", "c", metav1.NamespaceAll, "if specified read the KubeConfig for the management cluster from this context. Only applies when running out of cluster")
}

func rootCommand(ctx context.Context) *cobra.Command {
	var opts bootstrapOpts
	cmd := &cobra.Command{
		Use:     "cert-agent [command]",
		Short:   "Start the Service Mesh Hub Certificate Agent.",
		Long:    "The Service Mesh Hub Certificate Agent is used to generate certificates signed by Service Mesh Hub for use in managed clusters without requiring private keys to leave the managed cluster. For documentation on the actions taken by the Certificate Agent, see the generated documentation for the IssuedCertificate Custom Resource.",
		Version: version.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logrus.SetLevel(logrus.DebugLevel)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return startAgent(ctx, opts)
		},
	}

	opts.addToFlags(cmd.PersistentFlags())

	return cmd
}

func startAgent(ctx context.Context, opts bootstrapOpts) error {
	return agent.Start(ctx, opts.getBootstrap())
}
