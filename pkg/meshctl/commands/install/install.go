package install

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	cliversion "github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &options{GlobalFlags: globalFlags}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Gloo Mesh",
		Long: `Install the Gloo Mesh management plan to a Kubernetes cluster.

Go to https://www.solo.io/products/gloo-mesh/ to learn more about the
difference between the editions.
`,
	}

	cmd.AddCommand(
		communityCommand(ctx, opts),
		enterpriseCommand(ctx, opts),
	)

	opts.addToFlags(cmd.PersistentFlags())

	return cmd
}

type options struct {
	*utils.GlobalFlags
	dryRun bool

	kubeCfgPath     string
	kubeContext     string
	namespace       string
	chartPath       string
	chartValuesFile string
	version         string
	releaseName     string
	agentChartPath  string
	agentValuesPath string

	register      bool
	clusterName   string
	clusterDomain string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeCfgPath, &o.kubeContext, flags)
	flags.BoolVarP(&o.dryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace in which to install Gloo Mesh")
	flags.StringVar(&o.chartPath, "chart-file", "", "Path to a local Helm chart for installing Gloo Mesh.\nIf unset, this command will install Gloo Mesh from the publicly released Helm chart.")
	flags.StringVar(&o.chartValuesFile, "chart-values-file", "", "File containing value overrides for the Gloo Mesh Helm chart")
	flags.StringVar(&o.version, "version", "", "Version to install.\nCommunity defaults to meshctl version, enterprise defaults to latest stable")

	flags.BoolVarP(&o.register, "register", "r", false, "Also register the cluster")
	flags.StringVar(&o.clusterName, "cluster-name", "mgmt-cluster",
		"Name with which to register the cluster running Gloo Mesh, only applies if --register is also set")
	flags.StringVar(&o.clusterDomain, "cluster-domain", defaults.DefaultClusterDomain, "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. \nRead more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
}

func (o *options) getInstaller(chartUriTemplate string) helm.Installer {
	// User-specified chartPath takes precedence over specified version.
	if o.chartPath == "" {
		o.chartPath = fmt.Sprintf(chartUriTemplate, o.version)
	}

	logrus.Debugf("installing chart from %s", o.chartPath)
	return helm.Installer{
		ChartUri:    o.chartPath,
		ValuesFile:  o.chartValuesFile,
		KubeConfig:  o.kubeCfgPath,
		KubeContext: o.kubeContext,
		Namespace:   o.namespace,
		ReleaseName: o.releaseName,
		Values:      make(map[string]string),
		Verbose:     o.Verbose,
		DryRun:      o.dryRun,
	}
}

func (o *options) getRegistrationOptions() registration.Options {
	return registration.Options{
		KubeConfigPath:         o.kubeCfgPath,
		MgmtContext:            o.kubeContext,
		RemoteContext:          o.kubeContext,
		ClusterName:            o.clusterName,
		MgmtNamespace:          o.namespace,
		RemoteNamespace:        o.namespace,
		AgentChartPathOverride: o.agentChartPath,
		AgentChartValuesPath:   o.agentValuesPath,
		ClusterDomain:          o.clusterDomain,
		Version:                o.version,
		Verbose:                o.Verbose,
	}
}

func communityCommand(ctx context.Context, installOpts *options) *cobra.Command {
	opts := communityOptions{options: installOpts}
	cmd := &cobra.Command{
		Use:   "community",
		Short: "Install Gloo Mesh Community",
		Example: `  # Install to the currently selected Kubernetes context
  meshctl install community

  # Install to and register the currently selected Kubernetes context
  meshctl install community --register

  # Install to a different context
  meshctl install --kubecontext=some-context community`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installCommunity(ctx, opts)
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type communityOptions struct {
	*options
	agentCrdsChartPath string
	apiServerAddress   string
}

func (o *communityOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.releaseName, "release-name", gloomesh.GlooMeshReleaseName, "Helm release name")
	flags.StringVar(&o.apiServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag.\nSet this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
	flags.StringVar(&o.agentCrdsChartPath, "agent-crds-chart-file", "", "Path to a local Helm chart for installing CRDs needed by remote agents.\nIf unset, this command will install the agent CRDs from the publicly released Helm chart.")
	flags.StringVar(&o.agentChartPath, "cert-agent-chart-file", "",
		"Path to a local Helm chart for installing the Certificate Agent.\n"+
			"If unset, this command will install the Certificate Agent from the publicly released Helm chart.",
	)
	flags.StringVar(&o.agentValuesPath, "cert-agent-chart-values", "",
		"Path to a Helm values.yaml file for customizing the installation of the Certificate Agent.\n"+
			"If unset, this command will install the Certificate Agent with default Helm values.",
	)
}

func (o communityOptions) getInstaller() helm.Installer {
	return o.options.getInstaller(gloomesh.GlooMeshChartUriTemplate)
}

func (o communityOptions) getRegistrationOptions() registration.Options {
	reg := o.options.getRegistrationOptions()
	reg.AgentCrdsChartPath = o.agentCrdsChartPath
	reg.ApiServerAddress = o.apiServerAddress
	return reg
}

func installCommunity(ctx context.Context, opts communityOptions) error {
	const (
		repoURI   = "https://storage.googleapis.com/gloo-mesh"
		chartName = "gloo-mesh"
	)
	if opts.version == "" {
		opts.version = cliversion.Version
	}
	logrus.Info("Installing Helm chart")
	if err := opts.getInstaller().InstallChart(ctx); err != nil {
		return eris.Wrap(err, "installing gloo-mesh")
	}

	if opts.register && !opts.dryRun {
		logrus.Info("Registering cluster")
		registrant, err := registration.NewRegistrant(opts.getRegistrationOptions())
		if err != nil {
			return eris.Wrap(err, "initializing registrant")
		}
		if err := registrant.RegisterCluster(ctx); err != nil {
			return eris.Wrap(err, "registering management-plane cluster")
		}
	}

	return nil
}

func enterpriseCommand(ctx context.Context, installOpts *options) *cobra.Command {
	opts := enterpriseOptions{options: installOpts}
	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Install Gloo Mesh Enterprise (requires a license)",
		Example: `  # Install to the currently selected Kubernetes context
  meshctl install enterprise --license=<my_license>

  # Install to and register the currently selected Kubernetes context
  meshctl install enterprise --license=<my_license> --register

  # Don't install the UI
  meshctl install enterprise --license=<my_license> --skip-ui`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installEnterprise(ctx, opts)
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type enterpriseOptions struct {
	*options
	licenseKey         string
	skipUI             bool
	includeRBAC        bool
	relayServerAddress string
}

func (o *enterpriseOptions) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.releaseName, "release-name", gloomesh.GlooMeshReleaseName, "Helm release name")
	flags.StringVar(&o.licenseKey, "license", "", "Gloo Mesh Enterprise license key (required)")
	cobra.MarkFlagRequired(flags, "license")
	flags.BoolVar(&o.skipUI, "skip-ui", false, "Skip installation of the Gloo Mesh UI")
	flags.BoolVar(&o.includeRBAC, "include-rbac", false, "Install the RBAC Webhook")
	flags.StringVar(&o.agentChartPath, "enterprise-agent-chart-file", "",
		"Path to a local Helm chart for installing the Enterprise Agent.\n"+
			"If unset, this command will install the Enterprise Agent from the publicly released Helm chart.",
	)
	flags.StringVar(&o.agentValuesPath, "enterprise-agent-chart-values", "",
		"Path to a Helm values.yaml file for customizing the installation of the Enterprise Agent.\n"+
			"If unset, this command will install the Enterprise Agent with default Helm values.",
	)
	flags.StringVar(&o.relayServerAddress, "relay-server-address", "", "The address that the enterprise agentw will communicate with the relay server via.")
}

func (o enterpriseOptions) getInstaller() helm.Installer {
	ins := o.options.getInstaller(gloomesh.GlooMeshEnterpriseChartUriTemplate)
	ins.ReleaseName = o.releaseName
	ins.Values["licenseKey"] = o.licenseKey
	if o.skipUI {
		ins.Values["gloo-mesh-ui.enabled"] = "false"
	}
	if o.includeRBAC {
		ins.Values["rbac-webhook.enabled"] = "true"
	}

	return ins
}

func (o enterpriseOptions) getRegistrationOptions() enterprise.RegistrationOptions {
	if o.relayServerAddress == "" {
		const localRelayServerAddressFormat = "enterprise-networking.%s.svc.cluster.local:9900"
		namespacedLocalRelayServerAddress := fmt.Sprintf(localRelayServerAddressFormat, o.namespace)
		logrus.Infof("No relay server address provided, defaulting to %s", namespacedLocalRelayServerAddress)
		o.relayServerAddress = namespacedLocalRelayServerAddress
	}

	registrationOptions := enterprise.RegistrationOptions{
		Options:            o.options.getRegistrationOptions(),
		RelayServerAddress: o.relayServerAddress,
	}

	return registrationOptions
}

func installEnterprise(ctx context.Context, opts enterpriseOptions) error {
	const (
		repoURI   = "https://storage.googleapis.com/gloo-mesh-enterprise"
		chartName = "gloo-mesh-enterprise"
	)
	if opts.version == "" {
		cliVersion, err := version.NewVersion(cliversion.Version)
		if err != nil {
			return eris.Wrapf(err, "invalid CLI version: %s", cliversion.Version)
		}
		stable := cliVersion.Prerelease() == "" // Get latest stable if not using a pre-release CLI
		version, err := helm.GetLatestChartMinorVersion(
			repoURI, chartName, stable,
			cliVersion.Segments()[0], cliVersion.Segments()[1],
		)
		if err != nil {
			return err
		}
		opts.version = version
	}

	logrus.Info("Installing Helm chart")
	if err := opts.getInstaller().InstallChart(ctx); err != nil {
		return eris.Wrap(err, "installing gloo-mesh-enterprise")
	}
	if opts.register && !opts.dryRun {
		logrus.Info("Registering cluster")
		return enterprise.RegisterCluster(ctx, opts.getRegistrationOptions())
	}

	return nil
}
