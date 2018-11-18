package cmd

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/spf13/cobra"
)

var (
	filename          string
	overwriteExisting bool
	deleteExisting    bool
)

func App(version string) *cobra.Command {
	app := &cobra.Command{
		Use:   "supergloo",
		Short: "manage mesh resources with supergloo",
		Long: `superglooctl configures resources used by Supergloo server.
	Find more information at https://solo.io`,
		Version: version,
		// BashCompletionFunction: bashCompletion,
	}

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		dummySubCommand("sampleCmd"),
		dummySubCommand("anotherCmd"),
		install.Cmd(),
	)

	return app
}

func dummySubCommand(useName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   useName, // this should probably be hard coded in real use
		Short: `sample command`,
		Long: `long from sample command
long form sample continued.
TODO Replace with real content.
`,
		Run: func(c *cobra.Command, args []string) {
			fmt.Printf("you just ran cmd: %v\n", useName)
		},
	}
	pflags := cmd.PersistentFlags()
	pflags.StringVarP(&filename, "filename", "f", "", "filename to create resources from")
	pflags.BoolVarP(&overwriteExisting, "overwrite", "w", true, "overwrite existing resources "+
		"whose names overlap with those defined in the config file")
	pflags.BoolVarP(&deleteExisting, "delete", "d", false, "delete existing resources "+
		"whose names overlap with those defined in the config file")
	return cmd
}
