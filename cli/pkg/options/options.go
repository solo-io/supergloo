package options

import (
	"os"

	"github.com/spf13/pflag"
)

// wire provider func, not meant to be used outside of that
func NewOptionsProvider() *Options {
	return &Options{}
}

type Options struct {
	Root    Root
	Cluster Cluster
}

type Root struct {
	KubeConfig     string
	KubeContext    string
	WriteNamespace string
}

func AddRootFlags(pflags *pflag.FlagSet, options *Options) {
	pflags.StringVarP(&options.Root.WriteNamespace, "namespace", "n",
		"service-mesh-hub", "Specify the namespace which the resource should be written to")
	pflags.StringVar(&options.Root.KubeConfig, "kubeconfig",
		os.Getenv("KUBECONFIG"), "Specify the namespace which the resource should be written to")
	pflags.StringVar(&options.Root.KubeContext, "context", "",
		"Specify the context of the kube config which should be used, uses current context if none is specified")
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
