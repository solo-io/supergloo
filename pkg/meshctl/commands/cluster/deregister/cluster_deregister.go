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
		Long: `
Deregistering a cluster removes the installed agent from the remote cluster as
well as the other created resources such as service accounts. The edition
must match the edition that the cluster was originally registered with.

The name of the context of the cluster to dregister must be provided via the
--remote-context flag. It is important that the remote context and the name
passed as an argument are for the same cluster otherwise unexpected behavior
may occur.

If the management cluster is different than the one that the current context
points then it an be provided via the --mgmt-context flag.`,
		PersistentPreRun: func(*cobra.Command, []string) {
			opts.Verbose = globalFlags.Verbose
		},
	}

	cmd.AddCommand(
		communityCommand(ctx, opts),
		enterpriseCommand(ctx, opts),
	)

	opts.addToFlags(cmd.PersistentFlags())
	cmd.MarkFlagRequired("remote-context")

	return cmd
}

// Use type alias to allow defining receiver method in this package
type options registration.Options

func (o *options) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&o.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&o.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	set.StringVar(&o.MgmtKubeConfigPath, "mgmt-kubeconfig", "",
		"path to the kubeconfig file to use for the management cluster if different from control plane kubeconfig file location")
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
certificate agent and removing the cluster definition from the management cluster.`,
		Args:    cobra.MinimumNArgs(1),
		Example: " meshctl cluster deregister community remote-cluster --remote-context my-remote",
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
enterprise agent and removing the cluster definition from the management cluster.`,
		Args:    cobra.MinimumNArgs(1),
		Example: " meshctl cluster deregister enterprise remote-cluster --remote-context my-remote",
		RunE: func(_ *cobra.Command, args []string) error {
			opts.ClusterName = args[0]
			enterpriseOpts := enterprise.RegistrationOptions{Options: registration.Options(*opts)}
			return enterprise.DeregisterCluster(ctx, enterpriseOpts)
		},
	}

	cmd.SilenceUsage = true
	return cmd
}
