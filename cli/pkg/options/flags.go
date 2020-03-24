package options

import (
	"fmt"
	"os"
	"strings"

	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/spf13/cobra"
)

func AddRootFlags(cmd *cobra.Command, options *Options) {
	flags := cmd.PersistentFlags()
	flags.StringVarP(&options.Root.WriteNamespace, "namespace", "n", "service-mesh-hub",
		"Specify the namespace where Service Mesh Hub resources should be written")
	flags.StringVar(&options.Root.KubeConfig, "kubeconfig", os.Getenv("KUBECONFIG"),
		"Specify the kubeconfig for the current command")
	flags.StringVar(&options.Root.KubeContext, "context", "",
		"Specify which context from the kubeconfig should be used; uses current context if none is specified")
	flags.DurationVar(&options.Root.KubeTimeout, "kube-timeout", cliconstants.DefaultKubeClientTimeout,
		"Specify the default timeout for requests to kubernetes API servers.")
	flags.BoolVarP(&options.Root.Verbose, "verbose", "v", false,
		"Enable verbose mode, which outputs additional execution details that may be helpful for debugging")
}

func AddUpgradeFlags(cmd *cobra.Command, opts *Upgrade) {
	flags := cmd.PersistentFlags()
	flags.StringVar(&opts.ReleaseTag, "release", cliconstants.DefaultReleaseTag,
		"Which meshctl release to download. Specify a tag corresponding to the desired version of meshctl or \"latest\". "+
			"Service Mesh Hub releases can be found here: https://github.com/solo-io/service-mesh-hub/releases. "+
			"Omitting this tag defaults to \"latest\".")
	flags.StringVar(&opts.DownloadPath, "path", "",
		"Desired path for your upgraded meshctl binary. Defaults to the location of your currently executing binary.")
}

func AddInstallFlags(cmd *cobra.Command, opts *Options) {
	flags := cmd.Flags()
	flags.BoolVarP(&opts.SmhInstall.DryRun, "dry-run", "d", false,
		"Send the raw installation yaml to stdout instead of applying it to kubernetes")
	flags.StringVarP(&opts.SmhInstall.HelmChartOverride, "file", "f", "",
		"Install Service Mesh Hub from this Helm chart archive file rather than from a release")
	flags.StringSliceVarP(&opts.SmhInstall.HelmChartValueFileNames, "values", "", []string{},
		"List of files with value overrides for the Service Mesh Hub Helm chart, "+
			"(e.g. --values file1,file2 or --values file1 --values file2)")
	flags.StringVar(&opts.SmhInstall.HelmReleaseName, "release-name", cliconstants.ServiceMeshHubReleaseName,
		"Helm release name")
	flags.StringVar(&opts.SmhInstall.Version, "version", "",
		"Version to install (e.g. v1.2.0, defaults to latest)")
	flags.BoolVar(&opts.SmhInstall.CreateNamespace, "create-namespace", true,
		"Create the namespace to install Service Mesh Hub into")
}

const (
	ClusterRegisterOverwriteFlag = "overwrite"
)

func AddClusterRegisterFlags(cmd *cobra.Command, opts *Options) {
	flags := cmd.PersistentFlags()
	remoteClusterName := "remote-cluster-name"
	remoteWriteNamespace := "remote-write-namespace"
	remoteContext := "remote-context"
	remoteKubeconfig := "remote-kubeconfig"
	localClusterDomainOverride := "local-cluster-domain-override"
	devCsrAgentChartName := "dev-csr-agent-chart"

	flags.StringVar(&opts.Cluster.Register.RemoteClusterName, remoteClusterName, "",
		"Name of the cluster to be operated upon")
	flags.StringVar(&opts.Cluster.Register.RemoteWriteNamespace, remoteWriteNamespace, "default",
		"Namespace in the remote cluster in which to write resources")
	flags.StringVar(&opts.Cluster.Register.RemoteContext, remoteContext, "",
		"Set the context you would like to use for the remote cluster")
	flags.StringVar(&opts.Cluster.Register.RemoteKubeConfig, remoteKubeconfig, "",
		"Set the path to the kubeconfig you would like to use for the remote cluster. Leave empty to use the default.")
	flags.StringVar(&opts.Cluster.Register.LocalClusterDomainOverride, localClusterDomainOverride, "",
		"Swap out the domain of the remote cluster's k8s API server for the value of this flag; used mainly for debugging locally in docker, where you may provide a value like 'host.docker.internal'")
	flags.BoolVar(&opts.Cluster.Register.Overwrite, ClusterRegisterOverwriteFlag, false,
		"Overwrite any cluster registered with the cluster name provided")
	flags.BoolVar(&opts.Cluster.Register.UseDevCsrAgentChart, devCsrAgentChartName, false, "Use a packaged CSR agent chart from ./_output rather than a release chart")

	// this flag is mainly for our own debugging purposes
	// don't show it in usage messages
	flags.Lookup(localClusterDomainOverride).Hidden = true
	flags.Lookup(devCsrAgentChartName).Hidden = true

	cobra.MarkFlagRequired(flags, remoteClusterName)
}

func AddIstioInstallFlags(cmd *cobra.Command, opts *Options, profilesUsage string) {
	operatorNsFlag := "operator-namespace"
	flags := cmd.PersistentFlags()
	flags.StringVar(&opts.Istio.Install.InstallationConfig.IstioOperatorVersion, "operator-version", cliconstants.DefaultIstioOperatorVersion, "Version of the Istio operator to use (https://hub.docker.com/r/istio/operator/tags)")
	flags.StringVar(&opts.Istio.Install.InstallationConfig.InstallNamespace, operatorNsFlag, cliconstants.DefaultIstioOperatorNamespace, "Namespace in which to install the Istio operator")
	flags.BoolVar(&opts.Istio.Install.InstallationConfig.CreateIstioControlPlaneCRD, "create-operator-crd", true, "Register the IstioControlPlane CRD in the remote cluster")
	flags.BoolVar(&opts.Istio.Install.InstallationConfig.CreateNamespace, "create-operator-namespace", true, "Create the namespace specified by --"+operatorNsFlag)
	flags.BoolVar(&opts.Istio.Install.DryRun, "dry-run", false, "Dump the manifest that would be used to install the operator to stdout rather than apply it")
	flags.StringVar(&opts.Istio.Install.IstioControlPlaneManifestPath, "control-plane-spec", "", "Optional path to a YAML file containing an IstioControlPlane resource")
	flags.StringVar(&opts.Istio.Install.Profile, "profile", "", profilesUsage)
}

func AddUninstallFlags(cmd *cobra.Command, opts *Options) {
	flags := cmd.PersistentFlags()

	flags.StringVar(&opts.SmhUninstall.ReleaseName, "release-name", cliconstants.ServiceMeshHubReleaseName, "Helm release name")
	flags.BoolVar(&opts.SmhUninstall.RemoveNamespace, "remove-namespace", false, "Remove the Service Mesh Hub namespace specified with -n")
}

func AddCheckFlags(cmd *cobra.Command, opts *Options, defaultOutputFormat string, validOutputFormats []string) {
	flags := cmd.PersistentFlags()

	flags.StringVarP(&opts.Check.OutputFormat, "output", "o", defaultOutputFormat, fmt.Sprintf("Output format for the report. Valid values: [%s]", strings.Join(validOutputFormats, ", ")))
}
