package version

import (
	"io"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/spf13/cobra"
)

type VersionCommand *cobra.Command

var VersionSet = wire.NewSet(
	VersionCmd,
)

func VersionCmd(out io.Writer, clientsFactory common.ClientsFactory, opts *options.Options) VersionCommand {
	cmdName := "version"
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: "Display the version of meshctl and Service Mesh Hub server components",
		RunE: func(_ *cobra.Command, _ []string) error {
			return ReportVersion(out, clientsFactory, opts)
		},
	}

	return cmd
}
