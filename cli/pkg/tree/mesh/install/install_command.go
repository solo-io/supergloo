package mesh_install

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	istio1_5 "github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio1.5"
	istio1_6 "github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio1.6"
	"github.com/spf13/cobra"
)

type MeshInstallCommand *cobra.Command

var (
	MeshInstallProviderSet = wire.NewSet(
		MeshInstallRootCmd,
		istio1_5.NewIstio1_5InstallCmd,
		istio1_6.NewIstio1_6InstallCmd,
	)
)

func MeshInstallRootCmd(istio15Command istio1_5.Istio1_5Cmd, istio16Command istio1_6.Istio1_6Cmd) MeshInstallCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.MeshInstallCommand.Use,
		Short:   cliconstants.MeshInstallCommand.Short,
		Aliases: cliconstants.MeshInstallCommand.Aliases,
		RunE:    common.NonTerminalCommand(cliconstants.MeshInstallCommand.Use),
	}

	cmd.AddCommand(istio15Command, istio16Command)

	return cmd
}
