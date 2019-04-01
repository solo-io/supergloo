package get

import (
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "get information about supergloo objects",
	}

	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)

	cmd.AddCommand(getMeshIngressCmd(opts))

	return cmd
}
