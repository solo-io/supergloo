package install

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/codegen/helm"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	installhelm "github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Gloo Mesh",
	}

	cmd.AddCommand(
		communityCommand(ctx, globalFlags),
		enterpriseCommand(ctx, globalFlags),
	)

	return cmd
}

type options struct {
	verbose bool
	dryRun  bool

	kubeCfgPath     string
	kubeContext     string
	namespace       string
	chartPath       string
	chartValuesFile string
	releaseName     string
	version         string

	register           bool
	clusterName        string
	apiServerAddress   string
	clusterDomain      string
	agentCrdsChartPath string
	agentChartPath     string
	agentValuesPath    string
}

func (o *options) addToFlags(flags *pflag.FlagSet, agentName, agentFlagPrefix string) {
	utils.AddManagementKubeconfigFlags(&o.kubeCfgPath, &o.kubeContext, flags)
	flags.BoolVarP(&o.dryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace in which to install Gloo Mesh")
	flags.StringVar(&o.chartPath, "chart-file", "", "Path to a local Helm chart for installing Gloo Mesh. If unset, this command will install Gloo Mesh from the publicly released Helm chart.")
	flags.StringVar(&o.chartValuesFile, "chart-values-file", "", "File containing value overrides for the Gloo Mesh Helm chart")
	flags.StringVar(&o.releaseName, "release-name", helm.Chart.Data.Name, "Helm release name")
	flags.StringVar(&o.version, "version", "", "Version to install, defaults to latest if omitted")

	flags.BoolVarP(&o.register, "register", "r", false, "Register the cluster running Gloo Mesh")
	flags.StringVar(&o.clusterName, "cluster-name", "mgmt-cluster",
		"Name with which to register the cluster running Gloo Mesh, only applies if --register is also set")
	flags.StringVar(&o.apiServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
	flags.StringVar(&o.clusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	flags.StringVar(&o.agentCrdsChartPath, "agent-crds-chart-file", "", "Path to a local Helm chart for installing CRDs needed by remote agents. If unset, this command will install the agent CRDs from the publicly released Helm chart.")
	utils.AddAgentFlags(&o.agentChartPath, &o.agentValuesPath, flags, agentName, agentFlagPrefix)
}

func (o *options) getInstaller(chartUriTemplate string) installhelm.Installer {
	// User-specified chartPath takes precedence over specified version.
	if o.chartPath == "" {
		o.chartPath = fmt.Sprintf(chartUriTemplate, o.version)
	}

	return installhelm.Installer{
		ChartUri:    o.chartPath,
		ValuesFile:  o.chartValuesFile,
		KubeConfig:  o.kubeCfgPath,
		KubeContext: o.kubeContext,
		Namespace:   o.namespace,
		ReleaseName: o.releaseName,
		Values:      make(map[string]string),
		Verbose:     o.verbose,
		DryRun:      o.dryRun,
	}
}

func (o *options) getRegistrationOptions() registration.RegistrantOptions {
	return registration.RegistrantOptions{
		KubeConfigPath: o.kubeCfgPath,
		MgmtContext:    o.kubeContext,
		RemoteContext:  o.kubeContext,
		Registration: register.RegistrationOptions{
			ClusterName:      o.clusterName,
			RemoteCtx:        o.kubeContext,
			Namespace:        o.namespace,
			RemoteNamespace:  o.namespace,
			APIServerAddress: o.apiServerAddress,
			ClusterDomain:    o.clusterDomain,
		},
		AgentCrdsChartPath:     o.agentCrdsChartPath,
		AgentChartPathOverride: o.agentChartPath,
		AgentChartValues:       o.agentValuesPath,
		Verbose:                o.verbose,
	}
}

func communityCommand(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:   "community",
		Short: "Install the community certificate agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.verbose = globalFlags.Verbose
			return installCommunity(ctx, opts)
		},
	}

	opts.addToFlags(cmd.Flags(), "Certificate Agent", "cert-agent-")
	cmd.SilenceUsage = true
	return cmd
}

func installCommunity(ctx context.Context, opts options) error {
	const (
		repoURI   = "https://storage.googleapis.com/gloo-mesh"
		chartName = "gloo-mesh"
	)
	if opts.version == "" {
		version, err := installhelm.GetLatestChartVersion(repoURI, chartName)
		if err != nil {
			return err
		}
		opts.version = version
	}
	if err := opts.getInstaller(gloomesh.GlooMeshChartUriTemplate).InstallChart(ctx); err != nil {
		return eris.Wrap(err, "installing gloo-mesh")
	}

	if opts.register && !opts.dryRun {
		registrantOpts := opts.getRegistrationOptions()
		registrant, err := registration.NewRegistrant(
			registrantOpts,
			gloomesh.CertAgentReleaseName,
			gloomesh.CertAgentChartUriTemplate,
		)
		if err != nil {
			return eris.Wrap(err, "initializing registrant")
		}
		if err := registrant.RegisterCluster(ctx); err != nil {
			return eris.Wrap(err, "registering management-plane cluster")
		}
	}

	return nil
}

type enterpriseOptions struct {
	options
	licenseKey string
	skipUI     bool
	skipRBAC   bool
}

func (o *enterpriseOptions) addToFlags(flags *pflag.FlagSet) {
	o.options.addToFlags(flags, "Enterprise Agent", "enterprise-agent-")
	flags.StringVar(&o.licenseKey, "license", "", "Gloo Mesh Enterprise license key")
	cobra.MarkFlagRequired(flags, "license")
	flags.BoolVar(&o.skipUI, "skip-ui", false, "Skip installation of the Gloo Mesh UI")
	flags.BoolVar(&o.skipRBAC, "skip-rbac", false, "Skip installation of the RBAC Webhook")
}

func enterpriseCommand(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := enterpriseOptions{}
	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Install the enterprise agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.verbose = globalFlags.Verbose
			return installEnterprise(ctx, opts)
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

func installEnterprise(ctx context.Context, opts enterpriseOptions) error {
	const (
		repoURI   = "https://storage.googleapis.com/gloo-mesh-enterprise"
		chartName = "gloo-mesh-enterprise"
	)
	if opts.version == "" {
		version, err := installhelm.GetLatestChartVersion(repoURI, chartName)
		if err != nil {
			return err
		}
		opts.version = version
	}

	installer := opts.getInstaller(gloomesh.GlooMeshEnterpriseChartUriTemplate)
	installer.Values["licenseKey"] = opts.licenseKey
	if opts.skipUI {
		installer.Values["gloo-mesh-ui.enabled"] = "false"
	}
	if opts.skipRBAC {
		installer.Values["rbac-webhook.enabled"] = "false"
	}
	if err := installer.InstallChart(ctx); err != nil {
		return eris.Wrap(err, "installing gloo-mesh-enterprise")
	}
	if opts.register && !opts.dryRun {
		registrantOpts := opts.getRegistrationOptions()
		registrant, err := registration.NewRegistrant(
			registrantOpts,
			gloomesh.EnterpriseAgentReleaseName,
			gloomesh.EnterpriseAgentChartUriTemplate,
		)
		if err != nil {
			return eris.Wrap(err, "initializing registrant")
		}
		if err := registrant.RegisterCluster(ctx); err != nil {
			return eris.Wrap(err, "registering management-plane cluster")
		}
	}

	return nil
}
