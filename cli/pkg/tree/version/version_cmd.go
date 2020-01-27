package version

import (
	"io"

	"github.com/solo-io/mesh-projects/cli/pkg/common"

	"github.com/spf13/cobra"
)

func VersionCmd(out io.Writer, clientsFactory common.ClientsFactory, globalFlagConfig *common.GlobalFlagConfig) *cobra.Command {
	cmdName := "version"
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: "Display the version of meshctl and Service Mesh Hub server components",
		RunE: func(_ *cobra.Command, _ []string) error {
			return ReportVersion(out, clientsFactory, globalFlagConfig)
		},
	}

	return cmd
}
