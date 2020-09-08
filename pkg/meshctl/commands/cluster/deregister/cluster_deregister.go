package deregister

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/registration"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := deregistrationOptions{}
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Deregister a Kubernetes cluster from Service Mesh Hub, cleaning up any associated resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgmtKubeCfg, remoteKubeCfg, err := opts.kubeConfig.ConstructClientConfigs()
			if err != nil {
				return err
			}
			opts.registrant.KubeCfg = mgmtKubeCfg
			opts.registrant.RemoteKubeCfg = remoteKubeCfg
			// We need to explicitly pass the remote context because of this open issue: https://github.com/kubernetes/client-go/issues/735
			opts.registrant.RemoteCtx = opts.kubeConfig.RemoteContext
			return registration.NewRegistrant(&opts.registrant).DeregisterCluster(ctx)
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

// Use type alias to allow defining receiver method in this package
type deregistrationOptions struct {
	registrant registration.RegistrantOptions
	kubeConfig utils.MgmtRemoteKubeConfigOptions
}

func (opts *deregistrationOptions) addToFlags(set *pflag.FlagSet) {
	opts.kubeConfig.AddMgmtRemoteKubeConfigFlags(set)
	set.StringVar(&opts.registrant.ClusterName, "cluster-name", "", "name of the cluster to deregister")
	set.StringVar(&opts.registrant.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Service-Mesh-Hub control plane in which the secret for the deregistered cluster will be created")
	set.StringVar(&opts.registrant.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&opts.registrant.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	set.BoolVar(&opts.registrant.Verbose, "verbose", true, "enable/disable verbose logging during installation of cert-agent")
}
