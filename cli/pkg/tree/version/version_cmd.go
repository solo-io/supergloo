package version

import (
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/spf13/cobra"
)

type VersionCommand *cobra.Command

var VersionSet = wire.NewSet(
	VersionCmd,
)

func VersionCmd(out io.Writer, clientsFactory common.ClientsFactory, opts *options.Options) VersionCommand {
	cmd := &cobra.Command{
		Use:   cliconstants.VersionCommand.Use,
		Short: cliconstants.VersionCommand.Short,
		RunE: func(_ *cobra.Command, _ []string) error {
			return ReportVersion(out, clientsFactory, opts)
		},
	}
	return cmd
}
