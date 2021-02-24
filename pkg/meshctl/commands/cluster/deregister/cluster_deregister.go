package deregister

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Deregister a Kubernetes cluster from Gloo Mesh, cleaning up any associated resources",
	}

	cmd.AddCommand(
		communityCommand(ctx, globalFlags),
		enterpriseCommand(ctx, globalFlags),
	)

	return cmd
}

// Use type alias to allow defining receiver method in this package
type options registration.RegistrantOptions

func (o *options) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&o.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&o.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	set.StringVar(&o.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	set.StringVar(&o.Registration.ClusterName, "cluster-name", "", "name of the cluster to deregister")
	set.StringVar(&o.Registration.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Gloo Mesh control plane in which the secret for the deregistered cluster will be created")
	set.StringVar(&o.Registration.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&o.Registration.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
}

func communityCommand(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:   "community",
		Short: "Remove the community certificate agent",
		RunE: func(*cobra.Command, []string) error {
			opts.Verbose = globalFlags.Verbose
			registrant, err := registration.NewRegistrant(
				registration.RegistrantOptions(opts),
				gloomesh.CertAgentReleaseName,
				gloomesh.CertAgentChartUriTemplate,
			)
			if err != nil {
				return err
			}

			return registrant.DeregisterCluster(ctx)
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

func enterpriseCommand(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Remove the enterprise agent",
		RunE: func(*cobra.Command, []string) error {
			opts.Verbose = globalFlags.Verbose
			registrant, err := registration.NewRegistrant(
				registration.RegistrantOptions(opts),
				gloomesh.EnterpriseAgentReleaseName,
				gloomesh.EnterpriseAgentChartUriTemplate,
			)
			if err != nil {
				return err
			}

			return registrant.DeregisterCluster(ctx)
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}
