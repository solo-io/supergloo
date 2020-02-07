package options

import (
	"os"
	"time"

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
	Root    Root
	Cluster Cluster
	Upgrade Upgrade
}

type Root struct {
	KubeConfig     string
	KubeContext    string
	WriteNamespace string
	KubeTimeout    time.Duration
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

type Upgrade struct {
	ReleaseTag   string
	DownloadPath string
}
