package cluster

import (
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/spf13/cobra"
)

// config for the cluster command tree that ultimately comes from cobra flag values
type FlagConfig struct {
	GlobalFlagConfig     *common.GlobalFlagConfig
	TargetClusterName    string
	TargetWriteNamespace string
	TargetKubeConfigPath string
}

// returns a FlagConfig instance that will eventually be populated by flag values
func AddFlags(cmd *cobra.Command, flagConfig *FlagConfig) *FlagConfig {
	flags := cmd.PersistentFlags()

	flags.StringVar(&flagConfig.TargetClusterName, "target-cluster-name", "", "Name of the cluster to be operated upon")
	flags.StringVar(&flagConfig.TargetWriteNamespace, "target-write-namespace", "default", "Namespace in the target cluster in which to write resources")
	flags.StringVar(&flagConfig.TargetKubeConfigPath, "target-cluster-config", "", "Set the path to the kube config for your target cluster")

	cobra.MarkFlagRequired(flags, "target-cluster-config")
	cobra.MarkFlagRequired(flags, "target-cluster-name")

	return flagConfig
}
