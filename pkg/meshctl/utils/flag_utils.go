package utils

import (
	"github.com/spf13/pflag"
)

// Set kubeconfig path and context flags for the management cluster.
func AddManagementKubeconfigFlags(kubeconfig, kubecontext *string, flags *pflag.FlagSet) {
	flags.StringVar(kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the management cluster will be accessed")
	flags.StringVar(kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
}
