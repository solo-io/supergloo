package istio

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio/install"
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
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
		RunE:  cli_util.NonTerminalCommand(cmdName),
	}

	istio.AddCommand(
		istioInstallationCmd,
	)

	return istio
}
