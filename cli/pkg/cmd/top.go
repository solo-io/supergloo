package cmd

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/ingresstoolbox"
	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

var opts options.Options

func App(version string) *cobra.Command {
	app := &cobra.Command{
		Use:   "supergloo",
		Short: "manage mesh resources with supergloo",
		Long: `superglooctl configures resources used by Supergloo server.
	Find more information at https://solo.io`,
		Version: version,
		// BashCompletionFunction: bashCompletion,
	}
	pflags := app.PersistentFlags()
	pflags.BoolVarP(&opts.Top.Static, "static", "s", false, "disable interactive mode")

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		install.Cmd(&opts),
		meshtoolbox.FaultInjection(&opts),
		meshtoolbox.LoadBalancing(&opts),
		meshtoolbox.Retries(&opts),
		ingresstoolbox.FortifyIngress(&opts),
		ingresstoolbox.AddRoute(&opts),
	)

	return app
}
