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
	deregistrationOpts := deregistrationOptions{}
	kubeConfigOptions := utils.MgmtRemoteKubeConfigOptions{}
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Deregister a Kubernetes cluster from Service Mesh Hub, cleaning up any associated resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgmtKubeCfg, remoteKubeCfg := kubeConfigOptions.ConstructDiskKubeCfg()
			deregistrationOpts.KubeCfg = mgmtKubeCfg
			deregistrationOpts.RemoteKubeCfg = remoteKubeCfg
			options := registration.RegistrantOptions(deregistrationOpts)
			return registration.NewRegistrant(&options).DeregisterCluster(ctx)
		},
	}
	deregistrationOpts.addToFlags(cmd.Flags(), &kubeConfigOptions)
	cmd.SilenceUsage = true
	return cmd
}

// Use type alias to allow defining receiver method in this package
type deregistrationOptions registration.RegistrantOptions

func (opts *deregistrationOptions) addToFlags(set *pflag.FlagSet, kubeConfigOptions *utils.MgmtRemoteKubeConfigOptions) {
	utils.AddMgmtRemoteKubeConfigFlags(kubeConfigOptions, set)
	set.StringVar(&opts.ClusterName, "cluster-name", "", "name of the cluster to deregister")
	set.StringVar(&opts.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Service-Mesh-Hub control plane in which the secret for the deregistered cluster will be created")
	set.StringVar(&opts.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&opts.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	set.BoolVar(&opts.Verbose, "verbose", true, "enable/disable verbose logging during installation of cert-agent")
}
