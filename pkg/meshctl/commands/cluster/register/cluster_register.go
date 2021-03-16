package register

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
type options registration.Options

func (o *options) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.KubeConfigPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	flags.StringVar(&o.MgmtContext, "mgmt-context", "", "name of the kubeconfig context to use for the management cluster")
	flags.StringVar(&o.RemoteContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	flags.StringVar(&o.MgmtNamespace, "mgmt-namespace", defaults.DefaultPodNamespace, "namespace of the Gloo Mesh control plane in which the secret for the registered cluster will be created")
	flags.StringVar(&o.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created.\nIf the namespace does not exist it will be created.")
	flags.StringVar(&o.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'.\nRead more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
}

func communityCommand(ctx context.Context, regOpts *options) *cobra.Command {
	opts := communityOptions{Options: (*registration.Options)(regOpts)}
	cmd := &cobra.Command{
		Use:   "community [cluster name]",
		Short: "Register a cluster for Gloo Mesh community edition",
		Example: `  # Register the current cluster
  meshctl cluster register community mgmt-cluster

  # Register a different context when the current one is the management cluster
  meshctl cluster register --remote-context=my-context community remote-cluster`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			regOpts.AgentChartPathOverride = opts.AgentChartPathOverride
			regOpts.AgentChartValuesPath = opts.AgentChartValuesPath
			regOpts.ClusterName = args[0]
			registrant, err := registration.NewRegistrant(registration.Options(*regOpts))
			if err != nil {
				return err
			}

			logrus.Info("Registering cluster")
			return registrant.RegisterCluster(ctx)
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type communityOptions struct {
	*registration.Options

	AgentChartPathOverride string
	AgentChartValuesPath   string
}

func (o *communityOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.ApiServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag.\nSet this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
	flags.StringVar(&o.AgentCrdsChartPath, "agent-crds-chart-file", "", "Path to a local Helm chart for installing CRDs needed by remote agents.\nIf unset, this command will install the agent CRDs from the publicly released Helm chart.")
	flags.StringVar(&o.AgentChartPathOverride, "cert-agent-chart-file", "",
		"Path to a local Helm chart for installing the Certificate Agent.\n"+
			"If unset, this command will install the Certificate Agent from the publicly released Helm chart.",
	)
	flags.StringVar(&o.AgentChartValuesPath, "cert-agent-chart-values", "",
		"Path to a Helm values.yaml file for customizing the installation of the Certificate Agent.\n"+
			"If unset, this command will install the Certificate Agent with default Helm values.",
	)
}

func enterpriseCommand(ctx context.Context, regOpts *options) *cobra.Command {
	opts := enterpriseOptions{}
	cmd := &cobra.Command{
		Use:   "enterprise [cluster name]",
		Short: "Register a cluster for Gloo Mesh enterprise edition",
		Long: `In the process of registering a cluster, an agent (called the relay agent)
will be installed on the remote cluster. To establish trust between the relay agent and
Gloo-Mesh, mTLS is used.

The relay agent can either be provided with a client certificate, or a bootstrap token. If provided
with a bootstrap token, the relay agent will then exchange it for a client certificate and save it
as a secret in the cluster. Once the client certificate secret exist, the bootstrap token is not 
longer needed and can be discarded.

For the relay agent to trust Gloo Mesh a root CA is needed.

To make the registration process easy, this command will try to copy the root CA and 
bootstrap token from the gloo-mesh cluster, if these are not explicitly provided in the command line flags.
`,
		Example: `  # Register the current context
  meshctl cluster register enterprise mgmt-cluster

  # Register a different context when the current one is the management cluster
  meshctl cluster register --remote-context=my-context enterprise remote-cluster`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			logrus.Infof("Registering cluster")
			opts.Options = registration.Options(*regOpts)
			opts.ClusterName = args[0]
			return enterprise.RegisterCluster(ctx, enterprise.RegistrationOptions(opts))
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	cmd.MarkFlagRequired("relay-server-address")
	return cmd
}

type enterpriseOptions enterprise.RegistrationOptions

func (o *enterpriseOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.RelayServerAddress, "relay-server-address", "", "The address via which the enterprise agent will communicate with the relay server.")
	flags.BoolVar(&o.RelayServerInsecure, "relay-server-insecure", false, "Communicate with the relay server over an insecure connection.")

	flags.StringVar(&o.RootCASecretName, "root-ca-secret-name", "", "Secret name for the root CA for communication with relay server.")
	flags.StringVar(&o.RootCASecretNamespace, "root-ca-secret-namespace", "", "Secret namespace for the root CA for communication with relay server.")

	flags.StringVar(&o.ClientCertSecretName, "client-cert-secret-name", "", "Secret name for the client cert for communication with relay server. Note that if a bootstrap token is provided, then a client certificate will be created automatically.")
	flags.StringVar(&o.ClientCertSecretNamespace, "client-cert-secret-namespace", "", "Secret namespace for the client cert for communication with relay server.")

	flags.StringVar(&o.TokenSecretName, "token-secret-name", "", "Secret name for the bootstrap token. This token will be used to boostrap a client certificate from relay server. Not needed if already have a client certificate.")
	flags.StringVar(&o.TokenSecretNamespace, "token-secret-namespace", "", "Secret namespace for the bootstrap token.")
	flags.StringVar(&o.TokenSecretKey, "token-secret-key", "token", "Secret namespace for the bootstrap token.")

	flags.StringVar(&o.AgentChartPathOverride, "enterprise-agent-chart-file", "",
		"Path to a local Helm chart for installing the Enterprise Agent.\n"+
			"If unset, this command will install the Enterprise Agent from the publicly released Helm chart.",
	)
	flags.StringVar(&o.AgentChartValuesPath, "enterprise-agent-chart-values", "",
		"Path to a Helm values.yaml file for customizing the installation of the Enterprise Agent.\n"+
			"If unset, this command will install the Enterprise Agent with default Helm values.",
	)
}
