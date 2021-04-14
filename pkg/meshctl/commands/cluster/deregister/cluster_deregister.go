package deregister

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources",
		Long: `Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources

The edition deregistered must match the edition that was originally registered.`,
		PersistentPreRun: func(*cobra.Command, []string) {
			opts.Verbose = globalFlags.Verbose
		},
	}

	cmd.AddCommand(
		communityCommand(ctx, opts),
		enterpriseCommand(ctx, opts),
	)

	opts.addToFlags(cmd.PersistentFlags())

	return cmd
}

// Use type alias to allow defining receiver method in this package
type options registration.Options

func (o *options) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&o.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&o.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	set.StringVar(&o.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	set.StringVar(&o.ClusterName, "cluster-name", "", "name of the cluster to deregister")
	set.StringVar(&o.MgmtNamespace, "mgmt-namespace", defaults.DefaultPodNamespace, "namespace of the Gloo Mesh control plane in which the secret for the deregistered cluster will be created")
	set.StringVar(&o.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
}

type communityOptions options

func (o *communityOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.ApiServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
}

func communityCommand(ctx context.Context, deregOpts *options) *cobra.Command {
	opts := (*communityOptions)(deregOpts)
	cmd := &cobra.Command{
		Use:   "community [cluster name]",
		Short: "Remove the community certificate agent",
		Long: `Deregister the remote cluster, which includes uninstalling the
certificate agent and removing the cluster definition from the management cluster.

The name of the remote context must be passed in via the --remote-context flag
if it is different from the context currently selected in the kubernetes config.`,
		Args: cobra.MinimumNArgs(1),
		Example: `  # These examples assume that the currently selected context
  # is the management context.

  # Deregister the currently selected cluster
  # In this situation it is both the management and a remote cluster
  meshctl cluster deregister community mgmt-cluster

  # Deregister a different remote cluster separate from the current one
  meshctl cluster deregister community remote-cluster --remote-context my-remote-ctx`,
		RunE: func(_ *cobra.Command, args []string) error {
			opts.ClusterName = args[0]
			registrant, err := registration.NewRegistrant(registration.Options(*opts))
			if err != nil {
				return err
			}

			logrus.Infof("Deregistering cluster: %s", opts.ClusterName)
			return registrant.DeregisterCluster(ctx)
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

func enterpriseCommand(ctx context.Context, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enterprise [cluster name]",
		Short: "Remove the enterprise agent",
		Long: `Deregister the remote cluster, which includes uninstalling the
enterprise agent and removing the cluster definition from the management cluster.

The name of the remote context must be passed in via the --remote-context flag
if it is different from the context currently selected in the kubernetes config.`,
		Args: cobra.MinimumNArgs(1),
		Example: `  # These examples assume that the currently selected context
  # is the management context.

  # Deregister the currently selected cluster
  # In this situation it is both the management and a remote cluster
  meshctl cluster deregister enterprise mgmt-cluster

  # Deregister a different remote cluster separate from the current one
  meshctl cluster deregister enterprise remote-cluster --remote-context my-remote-ctx`,
		RunE: func(_ *cobra.Command, args []string) error {
			opts.ClusterName = args[0]
			enterpriseOpts := enterprise.RegistrationOptions{Options: registration.Options(*opts)}
			return enterprise.DeregisterCluster(ctx, enterpriseOpts)
		},
	}

	cmd.SilenceUsage = true
	return cmd
}
