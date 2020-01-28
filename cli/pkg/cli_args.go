package cli

import (
	"context"
	"io"
	"os"

	"github.com/solo-io/mesh-projects/pkg/env"

	"github.com/solo-io/mesh-projects/cli/pkg/common"
	clusterroot "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	"github.com/spf13/cobra"
)

type cliTree struct {
	cluster clusterTree
	version *cobra.Command
}

type clusterTree struct {
	rootCmd  *cobra.Command
	register *cobra.Command
}

// build an instance of the meshctl implementation
func BuildCli(ctx context.Context,
	clientsFactory common.ClientsFactory,
	out io.Writer,
	masterClusterVerifier common.MasterKubeConfigVerifier,
	usageReporter usageclient.Client,
) *cobra.Command {

	meshctl := &cobra.Command{
		Use:   "meshctl",
		Short: "CLI for Service Mesh Hub",
	}

	// set global flags
	globalFlagConfig := setGlobalFlags(meshctl, usageReporter, ctx, masterClusterVerifier)

	commandTree := buildCommandTree(clientsFactory, globalFlagConfig, out)

	// add cluster commands
	cluster := commandTree.cluster
	meshctl.AddCommand(cluster.rootCmd)
	cluster.rootCmd.AddCommand(cluster.register)

	// add version command
	meshctl.AddCommand(commandTree.version)

	return meshctl
}

func setGlobalFlags(cmd *cobra.Command,
	usageReporter usageclient.Client,
	ctx context.Context,
	masterClusterVerifier common.MasterKubeConfigVerifier) *common.GlobalFlagConfig {

	globalFlagConfig := &common.GlobalFlagConfig{}
	// Add the --master-cluster flag to the command, making it required if KUBECONFIG is unset.
	// Verify the supplied master cluster config.
	masterDefault := os.Getenv("KUBECONFIG")
	masterKubeConfigPath := ""
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		usageReporter.StartReportingUsage(ctx, common.UsageReportingInterval)
		verifiedMasterKubeConfig, err := masterClusterVerifier.Verify(&masterKubeConfigPath)
		if err != nil {
			return err
		}

		globalFlagConfig.MasterKubeConfig = verifiedMasterKubeConfig
		return nil
	}

	// master-write-namespace
	cmd.PersistentFlags().StringVar(
		&globalFlagConfig.MasterWriteNamespace,
		"master-write-namespace",
		env.DefaultWriteNamespace,
		"Set the namespace in which to write resources in your master cluster")
	// master-cluster-config
	cmd.PersistentFlags().StringVar(&masterKubeConfigPath,
		"master-cluster-config",
		masterDefault,
		"Set the path to the kube config for your master cluster, where Service Mesh Hub is installed")

	// if we couldn't pull a default from the env var, make the flag required
	if masterDefault == "" {
		cmd.MarkPersistentFlagRequired("master-cluster-config")
	}
	return globalFlagConfig
}

func buildCommandTree(clientsFactory common.ClientsFactory, globalFlagConfig *common.GlobalFlagConfig, out io.Writer) cliTree {
	return cliTree{
		cluster: clusterTree{
			rootCmd:  clusterroot.ClusterRootCmd(),
			register: register.ClusterRegistrationCmd(clientsFactory, globalFlagConfig, out),
		},
		version: version.VersionCmd(out, clientsFactory, globalFlagConfig),
	}
}
