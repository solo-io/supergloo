package cluster

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/register"
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
	"github.com/spf13/cobra"
)

type ClusterCommand *cobra.Command

var ClusterSet = wire.NewSet(
	register.RegistrationSet,
	ClusterRootCmd,
)

func ClusterRootCmd(registerCmd register.RegistrationCmd) ClusterCommand {
	cluster := &cobra.Command{
		Use:   cliconstants.ClusterCommand.Use,
		Short: cliconstants.ClusterCommand.Short,
		RunE:  cli_util.NonTerminalCommand(cliconstants.ClusterCommand.Use),
	}
	cluster.AddCommand(registerCmd)
	return cluster
}
