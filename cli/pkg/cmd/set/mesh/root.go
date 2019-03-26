package mesh

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mesh",
		Aliases: []string{"m"},
		Short:   "subcommands set a configuration parameters for one or more meshes",
	}

	cmd.AddCommand(setRootCertCmd(opts))
	cmd.AddCommand(setStatsCmd(opts))
	return cmd
}
