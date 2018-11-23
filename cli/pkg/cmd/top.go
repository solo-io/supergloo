package cmd

import (
	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/cmd/create"
	"github.com/solo-io/supergloo/cli/pkg/cmd/get"
	"github.com/solo-io/supergloo/cli/pkg/cmd/ingresstoolbox"
	"github.com/solo-io/supergloo/cli/pkg/cmd/initsupergloo"
	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/setup"
	"github.com/spf13/cobra"
)

var opts options.Options

func App(version string) *cobra.Command {
	app := &cobra.Command{
		Use:   "supergloo",
		Short: "manage mesh resources with supergloo",
		Long: `supergloo configures resources used by Supergloo server.
	Find more information at https://solo.io`,
		Version: version,
	}

	pflags := app.PersistentFlags()
	pflags.BoolVarP(&opts.Top.Static, "static", "s", false, "disable interactive mode")

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		initsupergloo.Cmd(&opts),
		install.Cmd(&opts),

		get.Cmd(&opts),
		create.Cmd(&opts),
		meshtoolbox.FaultInjection(&opts),
		meshtoolbox.LoadBalancing(&opts),
		meshtoolbox.Retries(&opts),
		ingresstoolbox.FortifyIngress(&opts),
		ingresstoolbox.AddRoute(&opts),
	)

	setup.InitCache(&opts)

	err := setup.InitSupergloo(&opts)
	if err != nil {
		panic(errors.Wrap(err, "Error during initialization."))
	}

	return app
}
