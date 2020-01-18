package version

import (
	"io"

	"github.com/spf13/cobra"
)

func VersionCmd(out io.Writer) *cobra.Command {
	cmdName := "version"
	version := &cobra.Command{
		Use:   cmdName,
		Short: "Display the version of meshctl",
		RunE: func(_ *cobra.Command, _ []string) error {
			return ReportVersion(out)
		},
	}

	return version
}
