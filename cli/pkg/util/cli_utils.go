package cli_util

import (
	"os"

	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/spf13/cobra"
)

// Add the --master-cluster flag to the command, making it required if KUBECONFIG is unset.
// This also causes master cluster verification to happen by setting up a prerun func on the command,
// passing along the callback we received.
func AddMasterFlag(cmd *cobra.Command, masterClusterVerifier common.MasterKubeConfigVerifier, onSuccess common.OnMasterVerificationSuccess) {
	flagSet := cmd.Flags()
	masterDefault := os.Getenv("KUBECONFIG")
	masterKubeConfigPath := ""

	flagSet.StringVar(&masterKubeConfigPath, "master-cluster-config", masterDefault, "Set the path to the kube config for your master cluster, where Service Mesh Hub is installed")

	cmd.PreRunE = masterClusterVerifier.BuildVerificationCallback(&masterKubeConfigPath, onSuccess)

	// if we couldn't pull a default from the env var, make the flag required
	if masterDefault == "" {
		cobra.MarkFlagRequired(flagSet, "master-cluster-config")
	}
}

// use when a non-terminal command is run directly, and without a subcommand- e.g. `meshctl cluster`
func NonTerminalCommand(commandName string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return eris.Errorf("Please provide a subcommand to `%s`", commandName)
	}
}
