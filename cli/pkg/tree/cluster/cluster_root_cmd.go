package cluster

import (
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
	"github.com/spf13/cobra"
)

// nothing to see here
func ClusterRootCmd() *cobra.Command {
	cmdName := "cluster"
	cluster := &cobra.Command{
		Use:   cmdName,
		Short: "Register and perform operations on clusters",
		RunE:  cli_util.NonTerminalCommand(cmdName),
	}

	return cluster
}
