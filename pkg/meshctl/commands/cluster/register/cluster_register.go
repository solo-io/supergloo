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
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Kubernetes cluster with Gloo Mesh",
		Long: `Register a Kubernetes cluster with Gloo Mesh

The edition registered must match the edition installed on the management cluster`,
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
type options registration.RegistrantOptions

func (o *options) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	flags.StringVar(&o.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	flags.StringVar(&o.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	flags.StringVar(&o.Registration.Namespace, "mgmt-namespace", defaults.DefaultPodNamespace, "namespace of the Gloo Mesh control plane in which the secret for the registered cluster will be created")
	flags.StringVar(&o.Registration.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created.\nIf the namespace does not exist it will be created.")
	flags.StringVar(&o.Registration.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag.\nSet this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
	flags.StringVar(&o.Registration.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'.\nRead more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	flags.StringVar(&o.AgentCrdsChartPath, "agent-crds-chart-file", "", "Path to a local Helm chart for installing CRDs needed by remote agents.\nIf unset, this command will install the agent CRDs from the publicly released Helm chart.")
}

func communityCommand(ctx context.Context, regOpts *options) *cobra.Command {
	opts := (*communityOptions)(regOpts)
	cmd := &cobra.Command{
		Use:   "community [cluster name]",
		Short: "Register using the community certificate agent",
		Example: `  # Register the current cluster
  meshctl cluster register community mgmt-cluster

  # Register a different context when the current one is the management cluster
  meshctl cluster register --remote-context=my-context community remote-cluster`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Registration.ClusterName = args[0]
			registrant, err := registration.NewRegistrant(
				registration.RegistrantOptions(*opts),
				gloomesh.CertAgentReleaseName,
				gloomesh.CertAgentChartUriTemplate,
			)
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

type communityOptions options

func (o *communityOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.AgentChartPathOverride, "cert-agent-chart-file", "",
		"Path to a local Helm chart for installing the Certificate Agent.\n"+
			"If unset, this command will install the Certificate Agent from the publicly released Helm chart.",
	)
	flags.StringVar(
		&o.AgentChartValues, "cert-agent-chart-values", "",
		"Path to a Helm values.yaml file for customizing the installation of the Certificate Agent.\n"+
			"If unset, this command will install the Certificate Agent with default Helm values.",
	)
}

func enterpriseCommand(ctx context.Context, regOpts *options) *cobra.Command {
	opts := (*enterpriseOptions)(regOpts)
	cmd := &cobra.Command{
		Use:   "enterprise [cluster name]",
		Short: "Register using the enterprise agent",
		Example: `  # Register the current context
  meshctl cluster register enterprise mgmt-cluster

  # Register a different context when the current one is the management cluster
  meshctl cluster register --remote-context=my-context enterprise remote-cluster`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Registration.ClusterName = args[0]
			registrant, err := registration.NewRegistrant(
				registration.RegistrantOptions(*opts),
				gloomesh.EnterpriseAgentReleaseName,
				gloomesh.EnterpriseAgentChartUriTemplate,
			)
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

type enterpriseOptions options

func (o *enterpriseOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.AgentChartPathOverride, "enterprise-agent-chart-file", "",
		"Path to a local Helm chart for installing the Enterprise Agent.\n"+
			"If unset, this command will install the Enterprise Agent from the publicly released Helm chart.",
	)
	flags.StringVar(
		&o.AgentChartValues, "enterprise-agent-chart-values", "",
		"Path to a Helm values.yaml file for customizing the installation of the Enterprise Agent.\n"+
			"If unset, this command will install the Enterprise Agent with default Helm values.",
	)
}
