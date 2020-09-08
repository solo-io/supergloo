package register

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/registration"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := registrationOptions{}
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Kubernetes cluster with Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			registrantOptions := registration.RegistrantOptions(opts)
			registrant, err := registration.NewRegistrant(&registrantOptions)
			if err != nil {
				return err
			}
			return registrant.RegisterCluster(ctx)
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

// Use type alias to allow defining receiver method in this package
type registrationOptions registration.RegistrantOptions

func (opts *registrationOptions) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&opts.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&opts.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	set.StringVar(&opts.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	set.StringVar(&opts.Registration.ClusterName, "cluster-name", "", "name of the cluster to register")
	set.StringVar(&opts.Registration.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Service-Mesh-Hub control plane in which the secret for the registered cluster will be created")
	set.StringVar(&opts.Registration.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&opts.Registration.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	set.StringVar(&opts.Registration.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	set.StringVar(&opts.CertAgent.ChartPath, "cert-agent-chart-file", "", "Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.")
	set.StringVar(&opts.CertAgent.ChartValues, "cert-agent-chart-values", "", "Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.")
	set.BoolVar(&opts.Verbose, "verbose", true, "enable/disable verbose logging during installation of cert-agent")
}
