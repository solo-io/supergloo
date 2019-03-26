package set

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/set/mesh"
	"github.com/solo-io/supergloo/cli/pkg/cmd/set/upstream"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{"s"},
		Short:   "update an existing resource with one or more config options",
	}

	cmd.AddCommand(
		mesh.Cmd(opts),
		upstream.Cmd(opts),
	)
	return cmd
}
