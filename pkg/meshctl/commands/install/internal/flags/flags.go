package flags

import (
	"fmt"

	"github.com/solo-io/gloo-mesh/codegen/helm"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/pflag"
)

type Options struct {
	KubeCfgPath     string
	KubeContext     string
	Namespace       string
	ChartPath       string
	CRDsChartPath   string
	ChartValuesFile string
	ReleaseName     string
	Version         string
	Verbose         bool
	DryRun          bool
	RegistrationOptions
}

type RegistrationOptions struct {
	Register            bool
	ClusterName         string
	ApiServerAddress    string
	ClusterDomain       string
	AgentCrdsChartPath  string
	CertAgentChartPath  string
	CertAgentValuesPath string
}

func (o *Options) AddToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.KubeCfgPath, &o.KubeContext, flags)
	flags.BoolVarP(&o.DryRun, "dry-run", "d", false, "Output installation manifest")
	flags.StringVar(&o.Namespace, "namespace", defaults.DefaultPodNamespace, "namespace in which to install Gloo Mesh")
	flags.StringVar(&o.ChartPath, "chart-file", "", "Path to a local Helm chart for installing Gloo Mesh. If unset, this command will install Gloo Mesh from the publicly released Helm chart.")
	flags.StringVar(&o.CRDsChartPath, "crds-chart-file", "", "Path to a local Helm chart for installing Gloo Mesh CRDs. If unset, this command will install Gloo Mesh CRDs from the publicly released Helm chart.")
	flags.StringVarP(&o.ChartValuesFile, "chart-values-file", "", "", "File containing value overrides for the Gloo Mesh Helm chart")
	flags.StringVar(&o.ReleaseName, "release-name", helm.Chart.Data.Name, "Helm release name")
	flags.StringVar(&o.Version, "version", "", "Version to install, defaults to latest if omitted")

	flags.BoolVarP(&o.Register, "register", "r", false, "Register the cluster running Gloo Mesh")
	flags.StringVar(&o.ClusterName, "cluster-name", "mgmt-cluster",
		"Name with which to register the cluster running Gloo Mesh, only applies if --register is also set")
	flags.StringVar(&o.ApiServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.")
	flags.StringVar(&o.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	flags.StringVar(&o.AgentCrdsChartPath, "agent-crds-chart-file", "", "Path to a local Helm chart for installing CRDs needed by remote agents. If unset, this command will install the agent CRDs from the publicly released Helm chart.")
	flags.StringVar(&o.CertAgentChartPath, "cert-agent-chart-file", "", "Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.")
	flags.StringVar(&o.CertAgentValuesPath, "cert-agent-chart-values", "", "Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.")
}

// TODO(ilackarms): clean up / decouple this constructor from the options
func (o *Options) GetInstaller(chartUriTemplate string, crds bool) gloomesh.Installer {
	// User-specified chartPath takes precedence over specified version.
	var chartPath string
	if !crds {
		chartPath = o.ChartPath
	} else {
		chartPath = o.CRDsChartPath
	}
	if chartPath == "" {
		chartPath = fmt.Sprintf(chartUriTemplate, o.Version)
	}

	return gloomesh.Installer{
		HelmChartPath:  chartPath,
		HelmValuesPath: o.ChartValuesFile,
		KubeConfig:     o.KubeCfgPath,
		KubeContext:    o.KubeContext,
		Namespace:      o.Namespace,
		ReleaseName:    o.ReleaseName,
		Values:         make(map[string]string),
		Verbose:        o.Verbose,
		DryRun:         o.DryRun,
	}
}

func (o *Options) GetRegistrationOptions() registration.RegistrantOptions {
	return registration.RegistrantOptions{
		KubeConfigPath: o.KubeCfgPath,
		MgmtContext:    o.KubeContext,
		RemoteContext:  o.KubeContext,
		Registration: register.RegistrationOptions{
			ClusterName:      o.ClusterName,
			RemoteCtx:        o.KubeContext,
			Namespace:        o.Namespace,
			RemoteNamespace:  o.Namespace,
			APIServerAddress: o.ApiServerAddress,
			ClusterDomain:    o.ClusterDomain,
		},
		AgentCrdsChartPath: o.AgentCrdsChartPath,
		CertAgent: registration.AgentInstallOptions{
			ChartPath:   o.CertAgentChartPath,
			ChartValues: o.CertAgentValuesPath,
		},
		Verbose: o.Verbose,
	}
}
