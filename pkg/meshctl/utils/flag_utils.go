package utils

import (
	"fmt"

	"github.com/spf13/pflag"
)

type GlobalFlags struct {
	Verbose bool
}

func (g *GlobalFlags) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&g.Verbose, "verbose", "v", false, "enable verbose logging")
}

// Set kubeconfig path and context flags for the management cluster.
func AddManagementKubeconfigFlags(kubeconfig, kubecontext *string, flags *pflag.FlagSet) {
	flags.StringVar(kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the management cluster will be accessed")
	flags.StringVar(kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
}

// AddAgentFlags adds the agent flags configured for the specific agent
func AddAgentFlags(chartPath, valuesPath *string, flags *pflag.FlagSet, agentName, flagPrefix string) {
	flags.StringVar(
		chartPath, flagPrefix+"chart-file", "",
		fmt.Sprintf(
			"Path to a local Helm chart for installing the %s. "+
				"If unset, this command will install the %s from the publicly released Helm chart.",
			agentName, agentName,
		),
	)
	flags.StringVar(
		valuesPath, flagPrefix+"chart-values", "",
		fmt.Sprintf(
			"Path to a Helm values.yaml file for customizing the installation of the %s. "+
				"If unset, this command will install the %s with default Helm values.",
			agentName, agentName,
		),
	)
}
