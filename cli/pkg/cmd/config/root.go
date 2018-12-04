package config

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/config/ca"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: `Configure mesh resources`,
		Long:  `Configure mesh resources`,
	}

	cmd.AddCommand(
		ca.Cmd(opts),
	)

	return cmd
}
