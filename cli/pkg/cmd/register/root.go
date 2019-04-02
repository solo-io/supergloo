package register

import (
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register",
		Aliases: []string{"r"},
		Short:   "commands for registering meshes with SuperGloo",
		Long:    "commands for registering meshes with SuperGloo",
	}

	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)

	cmd.AddCommand(registerAwsAppMeshCommand(opts))
	return cmd
}
