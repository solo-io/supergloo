package cluster

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/deregister"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register"
	"github.com/spf13/cobra"
)

type ClusterCommand *cobra.Command

var ClusterSet = wire.NewSet(
	register.RegistrationSet,
	deregister.DeregistrationSet,
	ClusterRootCmd,
)

func ClusterRootCmd(
	registerCmd register.RegistrationCmd,
	deregisterCmd deregister.DeregistrationCmd,
) ClusterCommand {
	cluster := &cobra.Command{
		Use:   cliconstants.ClusterCommand.Use,
		Short: cliconstants.ClusterCommand.Short,
		RunE:  common.NonTerminalCommand(cliconstants.ClusterCommand.Use),
	}
	cluster.AddCommand(
		registerCmd,
		deregisterCmd,
	)
	return cluster
}
