package istio

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/install"
	"github.com/spf13/cobra"
)

type IstioCommand *cobra.Command

var IstioProviderSet = wire.NewSet(
	install.IstioInstallationProviderSet,
	IstioRootCmd,
)

func IstioRootCmd(istioInstallationCmd install.IstioInstallationCmd, opts *options.Options) IstioCommand {
	cmdName := "istio"
	istio := &cobra.Command{
		Use:   cmdName,
		Short: "Manage installations of Istio",
		RunE:  common.NonTerminalCommand(cmdName),
	}

	istio.AddCommand(
		istioInstallationCmd,
	)

	return istio
}
