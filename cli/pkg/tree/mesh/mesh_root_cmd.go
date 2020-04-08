package mesh

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mesh_install "github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install"
	"github.com/spf13/cobra"
)

type MeshCommand *cobra.Command

var MeshProviderSet = wire.NewSet(
	mesh_install.MeshInstallProviderSet,
	MeshRootCmd,
)

func MeshRootCmd(meshInstallCommand mesh_install.MeshInstallCommand) MeshCommand {
	istio := &cobra.Command{
		Use:   cliconstants.MeshCommand.Use,
		Short: cliconstants.MeshCommand.Short,
		RunE:  common.NonTerminalCommand(cliconstants.MeshCommand.Use),
	}

	istio.AddCommand(
		meshInstallCommand,
	)

	return istio
}
