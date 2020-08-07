package check

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/check/internal"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Perform health checks on the Service Mesh Hub system",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			return internal.RunChecks(ctx, client, opts.namespace)
		},
	}
	opts.addToFlags(cmd.Flags())

	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Service Mesh Hub is installed in")
}
