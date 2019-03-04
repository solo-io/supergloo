package create

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "create a rule to apply to a mesh",
		Long: `Creates Rule resources which the SuperGloo controller 
will use to configure an installed mesh.`,
	}

	cmd.AddCommand(createRoutingRuleCmd(opts))
	return cmd
}
