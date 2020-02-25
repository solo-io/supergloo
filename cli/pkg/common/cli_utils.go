package common

import (
	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
)

// use when a non-terminal command is run directly, and without a subcommand- e.g. `meshctl cluster`
func NonTerminalCommand(commandName string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return eris.Errorf("Please provide a subcommand to `%s`", commandName)
	}
}
