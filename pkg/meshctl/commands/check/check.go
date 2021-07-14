package check

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/checks"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Perform health checks on the Gloo Mesh system",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeClient, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			return runChecks(ctx, kubeClient, opts)
		},
	}
	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
	localPort   uint32
	remotePort  uint32
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Gloo Mesh is installed in")
	flags.Uint32Var(&o.localPort, "local-port", defaults.MetricsPort, "local port used to open port-forward to enterprise mgmt pod (enterprise only)")
	flags.Uint32Var(&o.remotePort, "remote-port", defaults.MetricsPort, "remote port used to open port-forward to enterprise mgmt pod (enterprise only). set to 0 to disable checks on the mgmt server")
}

func runChecks(ctx context.Context, client client.Client, opts *options) error {
	checkCtx := checks.NewOutOfClusterCheckContext(client, opts.namespace, opts.kubeconfig, opts.kubecontext, opts.localPort, opts.remotePort)
	checks.RunChecks(ctx, checkCtx, checks.Server, checks.PostInstall)
	return nil
}
