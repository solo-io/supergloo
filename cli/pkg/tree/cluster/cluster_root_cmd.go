package cluster

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register"
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
	"github.com/spf13/cobra"
)

type ClusterCommand *cobra.Command

var ClusterSet = wire.NewSet(
	register.RegistrationSet,
	ClusterRootCmd,
)

func ClusterRootCmd(registerCmd register.RegistrationCmd, opts *options.Options) ClusterCommand {
	cmdName := "cluster"
	cluster := &cobra.Command{
		Use:   cmdName,
		Short: "Register and perform operations on clusters",
		RunE:  cli_util.NonTerminalCommand(cmdName),
	}
	cluster.AddCommand(registerCmd)

	return cluster
}
