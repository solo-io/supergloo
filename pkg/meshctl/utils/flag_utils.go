package utils

import (
	"github.com/spf13/pflag"
)

type GlobalFlags struct {
	Verbose bool
}

func (g *GlobalFlags) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&g.Verbose, "verbose", "v", false, "Enable verbose logging")
}

// Set kubeconfig path and context flags for the management cluster.
func AddManagementKubeconfigFlags(kubeconfig, kubecontext *string, flags *pflag.FlagSet) {
	flags.StringVar(kubeconfig, "kubeconfig", "", "Path to the kubeconfig from which the management cluster will be accessed")
	flags.StringVar(kubecontext, "kubecontext", "", "Name of the kubeconfig context to use for the management cluster")
}

// Set kubeconfig path and context flags for any cluster.
func AddKubeconfigFlags(kubeconfig, kubecontext *string, flags *pflag.FlagSet) {
	flags.StringVar(kubeconfig, "kubeconfig", "", "Path to the kubeconfig from which the cluster will be accessed")
	flags.StringVar(kubecontext, "kubecontext", "", "Name of the kubeconfig context to use for the cluster")
}
