package common

import (
	"os"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
)

// use when a non-terminal command is run directly, and without a subcommand- e.g. `meshctl cluster`
func NonTerminalCommand(commandName string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return eris.Errorf("Please provide a subcommand to `%s`", commandName)
	}
}

func GetBinaryName(cmd *cobra.Command) string {
	// default to os.Args[0] to try to get a more accurate picture of how the user's binary is named
	// but if we're in a test bed where the binary that was invoked is not meshctl, just use the root command name
	// (which should just be meshctl)
	binaryName := os.Args[0]
	if !strings.Contains(binaryName, "meshctl") {
		binaryName = cmd.Root().Name()
	}
	return binaryName
}
