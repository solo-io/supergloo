package cluster

import (
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

// config for the cluster command tree that ultimately comes from cobra flag values
type FlagConfig struct {
	TargetClusterName    string
	TargetWriteNamespace string
	TargetKubeConfigPath string
	MasterWriteNamespace string
	MasterKubeConfig     *rest.Config
}

// returns a FlagConfig instance that will eventually be populated by flag values
func AddFlags(cmd *cobra.Command, masterClusterVerifier common.MasterKubeConfigVerifier) *FlagConfig {
	flags := cmd.PersistentFlags()

	flagConfig := &FlagConfig{}

	cli_util.AddMasterFlag(cmd, masterClusterVerifier, func(verifiedMasterKubeConfig *rest.Config) {
		// grab the parsed kube config from the verifier if the config is validated successfully
		flagConfig.MasterKubeConfig = verifiedMasterKubeConfig
	})

	flags.StringVar(&flagConfig.TargetClusterName, "target-cluster-name", "", "Name of the cluster to be operated upon")
	flags.StringVar(&flagConfig.TargetWriteNamespace, "target-write-namespace", "default", "Namespace in the target cluster in which to write resources")
	flags.StringVar(&flagConfig.TargetKubeConfigPath, "target-cluster-config", "", "Set the path to the kube config for your target cluster")
	flags.StringVar(&flagConfig.MasterWriteNamespace, "master-write-namespace", env.DefaultWriteNamespace, "Set the namespace in which to write resources in your master cluster")

	cobra.MarkFlagRequired(flags, "target-cluster-config")
	cobra.MarkFlagRequired(flags, "target-cluster-name")

	return flagConfig
}
