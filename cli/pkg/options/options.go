package options

import (
	"os"
	"time"

	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/operator/install"
	"github.com/spf13/pflag"
)

// wire provider func, not meant to be used outside of that
func NewOptionsProvider() *Options {
	return &Options{}
}

const (
	defaultKubeClientTimeout = 5 * time.Second
)

type Options struct {
	Root       Root
	Cluster    Cluster
	Upgrade    Upgrade
	SmhInstall SmhInstall
	Istio      Istio
}

type Root struct {
	KubeConfig     string
	KubeContext    string
	WriteNamespace string
	KubeTimeout    time.Duration
	Verbose        bool
}

func AddRootFlags(pflags *pflag.FlagSet, options *Options) {
	pflags.StringVarP(&options.Root.WriteNamespace, "namespace", "n",
		"service-mesh-hub", "Specify the namespace which the resource should be written to")
	pflags.StringVar(&options.Root.KubeConfig, "kubeconfig",
		os.Getenv("KUBECONFIG"), "Specify the namespace which the resource should be written to")
	pflags.StringVar(&options.Root.KubeContext, "context", "",
		"Specify the context of the kube config which should be used, uses current context if none is specified")
	pflags.DurationVar(&options.Root.KubeTimeout, "kube-timeout", defaultKubeClientTimeout, "Specify the default "+
		"timeout for requests to kubernetes API servers.")
	pflags.BoolVarP(&options.Root.Verbose, "verbose", "v", false,
		"Enable verbose mode, which outputs additional execution details that may be helpful for debugging")
}

type Cluster struct {
	Register Register
}

type Register struct {
	RemoteClusterName    string
	RemoteWriteNamespace string
	RemoteContext        string
	RemoteKubeConfig     string
}

type Istio struct {
	Install IstioInstall
}

type IstioInstall struct {
	InstallationConfig            install.InstallationConfig
	DryRun                        bool
	IstioControlPlaneManifestPath string
	Profile                       string
}

type Upgrade struct {
	ReleaseTag   string
	DownloadPath string
}

type SmhInstall struct {
	DryRun                  bool
	HelmChartOverride       string
	HelmChartValueFileNames []string
	HelmReleaseName         string
	Version                 string
	CreateNamespace         bool
}
