package cli

import (
	"io"

	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
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
func BuildCli(clientsFactory common.ClientsFactory, out io.Writer, masterClusterVerifier common.MasterKubeConfigVerifier) *cobra.Command {
	meshctl := &cobra.Command{
		Use:   "meshctl",
		Short: "CLI for Service Mesh Hub",
	}

	commandTree := buildCommandTree(clientsFactory, masterClusterVerifier, out)

	// add cluster commands
	cluster := commandTree.cluster
	meshctl.AddCommand(cluster.rootCmd)
	cluster.rootCmd.AddCommand(cluster.register)

	// add version command
	meshctl.AddCommand(commandTree.version)

	return meshctl
}

func buildCommandTree(clientsFactory common.ClientsFactory, masterClusterVerifier common.MasterKubeConfigVerifier, out io.Writer) cliTree {
	return cliTree{
		cluster: clusterTree{
			rootCmd:  cluster.ClusterRootCmd(),
			register: register.ClusterRegistrationCmd(clientsFactory, masterClusterVerifier, out),
		},
		version: version.VersionCmd(out),
	}
}
