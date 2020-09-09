package deregister

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/registration"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := deregistrationOptions{}
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Deregister a Kubernetes cluster from Service Mesh Hub, cleaning up any associated resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			registrantOptions := registration.RegistrantOptions(opts)
			registrant, err := registration.NewRegistrant(&registrantOptions)
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

// Use type alias to allow defining receiver method in this package
type deregistrationOptions registration.RegistrantOptions

func (opts *deregistrationOptions) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&opts.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&opts.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	set.StringVar(&opts.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	set.StringVar(&opts.Registration.ClusterName, "cluster-name", "", "name of the cluster to deregister")
	set.StringVar(&opts.Registration.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Service-Mesh-Hub control plane in which the secret for the deregistered cluster will be created")
	set.StringVar(&opts.Registration.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&opts.Registration.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	set.BoolVar(&opts.Verbose, "verbose", true, "enable/disable verbose logging during installation of cert-agent")
}
