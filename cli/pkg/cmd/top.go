package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/cmd/config"
	"github.com/solo-io/supergloo/cli/pkg/cmd/create"
	"github.com/solo-io/supergloo/cli/pkg/cmd/get"
	"github.com/solo-io/supergloo/cli/pkg/cmd/ingresstoolbox"
	"github.com/solo-io/supergloo/cli/pkg/cmd/initsupergloo"
	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/cmd/uninstall"
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

	pFlags := app.PersistentFlags()
	pFlags.BoolVarP(&opts.Top.Static, "static", "s", false, "disable interactive mode")

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		initsupergloo.Cmd(&opts),
		install.Cmd(&opts),
		uninstall.Cmd(&opts),

		get.Cmd(&opts),
		create.Cmd(&opts),
		config.Cmd(&opts),
		meshtoolbox.FaultInjection(&opts),
		meshtoolbox.LoadBalancing(&opts),
		meshtoolbox.Retries(&opts),
		meshtoolbox.Policy(&opts),
		meshtoolbox.ToggleMtls(&opts),
		ingresstoolbox.FortifyIngress(&opts),
		ingresstoolbox.AddRoute(&opts),
	)

	// Fail fast if we cannot connect to kubernetes
	err := setup.CheckConnection()
	if err != nil {
		fmt.Println(errors.Wrap(err, "Failed to connect to Kubernetes. Please check whether the current-context "+
			"in your kubeconfig file points to a running cluster"))
		os.Exit(1)
	}

	err = setup.InitCache(&opts)
	if err != nil {
		fmt.Println(errors.Wrap(err, "Error during initialization!"))
		os.Exit(1)
	}

	err = setup.InitSupergloo(&opts)
	if err != nil {
		fmt.Println(errors.Wrap(err, "Error during initialization!"))
		os.Exit(1)
	}

	return app
}
