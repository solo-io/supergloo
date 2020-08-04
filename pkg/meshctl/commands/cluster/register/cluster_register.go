package register

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/registration"

	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	registrant := &registration.Registrant{}
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Kubernetes cluster with Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return registrant.RegisterCluster(ctx)
		},
	}
	addToFlags(registrant, cmd.Flags())
	return cmd
}

func addToFlags(opts *registration.Registrant, set *pflag.FlagSet) {
	set.StringVar(&opts.ClusterName, "cluster-name", "", "name of the cluster to register")
	set.StringVar(&opts.KubeCfgPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&opts.KubeContext, "master-context", "", "name of the kubeconfig context to use for the master cluster")
	set.StringVar(&opts.RemoteKubeContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	set.StringVar(&opts.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Service-Mesh-Hub control plane in which the secret for the registered cluster will be created")
	set.StringVar(&opts.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&opts.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	set.StringVar(&opts.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	set.StringVar(&opts.CertAgentInstallOptions.ChartPath, "cert-agent-chart-file", "", "Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.")
	set.StringVar(&opts.CertAgentInstallOptions.ChartValues, "cert-agent-chart-values", "", "Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.")
	set.BoolVar(&opts.Verbose, "verbose", true, "enable/disable verbose logging during installation of cert-agent")
}
