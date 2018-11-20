package create

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: `Create a resource from stdin`,
		Long:  `Create a resource from stdin`,
		Args:  cobra.ExactArgs(1), // TODO: for now allow only stdin creation, no file
		//DisableFlagsInUseLine: true,
		Run: func(c *cobra.Command, args []string) {
			// TODO: do nothing for now, we'll handle the -f option here when we add it
		},
	}

	cmd.AddCommand(RoutingRuleCmd(opts))

	return cmd
}
