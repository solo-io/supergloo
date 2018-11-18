package cmd

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/spf13/cobra"
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
		install.Cmd(),
	)

	return app
}
