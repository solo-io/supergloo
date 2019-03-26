package set

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{"s"},
		Short:   "subcommands set a configuration parameter for one or more meshes",
	}

	cmd.AddCommand(setRootCertCmd(opts))
	cmd.AddCommand(setStatsCmd(opts))
	return cmd
}
