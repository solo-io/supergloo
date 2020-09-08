package install

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/codegen/helm"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/registration"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeCfgPath     string
	kubeContext     string
	namespace       string
	chartPath       string
	chartValuesFile string
	releaseName     string
	version         string
	verbose         bool
	dryRun          bool
	registrationOptions
}

type registrationOptions struct {
	register            bool
	clusterName         string
	apiServerAddress    string
	clusterDomain       string
	certAgentChartPath  string
	certAgentValuesPath string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeCfgPath, &o.kubeContext, flags)
	flags.BoolVarP(&o.dryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace in which to install Service Mesh Hub")
	flags.StringVar(&o.chartPath, "chart-file", "", "Path to a local Helm chart for installing Service Mesh Hub. If unset, this command will install Service Mesh Hub from the publicly released Helm chart.")
	flags.StringVarP(&o.chartValuesFile, "chart-values-file", "", "", "File containing value overrides for the Service Mesh Hub Helm chart")
	flags.StringVar(&o.releaseName, "release-name", helm.Chart.Data.Name, "Helm release name")
	flags.StringVar(&o.version, "version", "", "Version to install, defaults to latest if omitted")

	flags.BoolVarP(&o.register, "register", "r", false, "Register the cluster running Service Mesh Hub")
	flags.StringVar(&o.clusterName, "cluster-name", "mgmt-cluster",
		"Name with which to register the cluster running Service Mesh Hub, only applies if --register is also set")
	flags.StringVar(&o.apiServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	flags.StringVar(&o.clusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	flags.StringVar(&o.certAgentChartPath, "cert-agent-chart-file", "", "Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.")
	flags.StringVar(&o.certAgentValuesPath, "cert-agent-chart-values", "", "Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.")
	flags.BoolVarP(&o.verbose, "verbose", "v", false, "Enable verbose output")
}

func install(ctx context.Context, opts *options) error {
	// User-specified chartPath takes precedence over specified version.
	smhChartUri := opts.chartPath
	smhVersion := opts.version
	if opts.version == "" {
		smhVersion = version.Version
	}
	if smhChartUri == "" {
		smhChartUri = fmt.Sprintf(smh.ServiceMeshHubChartUriTemplate, smhVersion)
	}

	err := smh.Installer{
		HelmChartPath:  smhChartUri,
		HelmValuesPath: opts.chartValuesFile,
		KubeConfig:     opts.kubeCfgPath,
		KubeContext:    opts.kubeContext,
		Namespace:      opts.namespace,
		ReleaseName:    opts.releaseName,
		Verbose:        opts.verbose,
		DryRun:         opts.dryRun,
	}.InstallServiceMeshHub(
		ctx,
	)

	if err != nil {
		return eris.Wrap(err, "installing service-mesh-hub")
	}

	if opts.register && !opts.dryRun {
		registrantOpts := &registration.RegistrantOptions{
			KubeConfigPath: opts.kubeCfgPath,
			MgmtContext:    opts.kubeContext,
			RemoteContext:  opts.kubeContext,
			Registration: register.RegistrationOptions{
				ClusterName:      opts.clusterName,
				RemoteCtx:        opts.kubeContext,
				Namespace:        opts.namespace,
				RemoteNamespace:  opts.namespace,
				APIServerAddress: opts.apiServerAddress,
				ClusterDomain:    opts.clusterDomain,
			},
			CertAgent: registration.CertAgentInstallOptions{
				ChartPath:   opts.certAgentChartPath,
				ChartValues: opts.certAgentValuesPath,
			},
			Verbose: opts.verbose,
		}
		registrant, err := registration.NewRegistrant(registrantOpts)
		if err != nil {
			return eris.Wrap(err, "initializing registrant")
		}
		if err := registrant.RegisterCluster(ctx); err != nil {
			return eris.Wrap(err, "registering management-plane cluster")
		}
	}
	return nil
}
