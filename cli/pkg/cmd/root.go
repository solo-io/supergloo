package cmd

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

var opts options.Options

func SuperglooCli(version string) *cobra.Command {
	app := &cobra.Command{
		Use:   "supergloo",
		Short: "CLI for Supergloo",
		Long: `supergloo configures resources watched by the Supergloo Controller.
	Find more information at https://solo.io`,
		Version: version,
	}

	pflags := app.PersistentFlags()
	pflags.BoolVarP(&opts.Interactive, "interactive", "i", false, "use interactive mode")

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		install.Cmd(&opts),
	)

	return app
}
