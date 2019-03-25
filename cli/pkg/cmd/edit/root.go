package edit

import (
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit",
		Aliases: []string{"e"},
		Short:   "commands for editing resources used by SuperGloo",
		Long:    "commands for editing resources used by SuperGloo",
	}

	cmd.AddCommand(editUpstreamCommand(opts))
	return cmd
}

func editUpstreamCommand(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upstream",
		Aliases: []string{"u"},
		Short:   "edit a gloo upstream for use within SuperGloo.",
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)

	cmd.AddCommand(tlsCmd(opts))
	cmd.AddCommand(awsCmd(opts))
	return cmd
}
