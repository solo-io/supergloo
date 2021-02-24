package register

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
		Use:   "register",
		Short: "Register a Kubernetes cluster with Gloo Mesh",
	}

	cmd.AddCommand(
		communityCommand(ctx, globalFlags),
		enterpriseCommand(ctx, globalFlags),
	)

	return cmd
}

// Use type alias to allow defining receiver method in this package
type options registration.RegistrantOptions

func (o *options) addToFlags(flags *pflag.FlagSet, agentName, agentFlagPrefix string) {
	flags.StringVar(&o.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	flags.StringVar(&o.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	flags.StringVar(&o.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	flags.StringVar(&o.Registration.ClusterName, "cluster-name", "", "name of the cluster to register")
	flags.StringVar(&o.Registration.Namespace, "mgmt-namespace", defaults.DefaultPodNamespace, "namespace of the Gloo Mesh control plane in which the secret for the registered cluster will be created")
	flags.StringVar(&o.Registration.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	flags.StringVar(&o.Registration.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
	flags.StringVar(&o.Registration.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	flags.StringVar(&o.AgentCrdsChartPath, "agent-crds-chart-file", "", "Path to a local Helm chart for installing CRDs needed by remote agents. If unset, this command will install the agent CRDs from the publicly released Helm chart.")
	utils.AddAgentFlags(&o.AgentChartPathOverride, &o.AgentChartValues, flags, agentName, agentFlagPrefix)
}

func communityCommand(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:   "community",
		Short: "Register using the community certificate agent",
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

			return registrant.RegisterCluster(ctx)
		},
	}

	opts.addToFlags(cmd.Flags(), "Certificate Agent", "cert-agent-")
	cmd.SilenceUsage = true
	return cmd
}

func enterpriseCommand(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Register using the enterprise agent",
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

			return registrant.RegisterCluster(ctx)
		},
	}

	opts.addToFlags(cmd.Flags(), "Enterprise Agent", "enterprise-agent-")
	cmd.SilenceUsage = true
	return cmd
}
