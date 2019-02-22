package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/setup"
	"github.com/spf13/cobra"
)

var opts options.Options

func SuperglooCli(version string) *cobra.Command {
	app := &cobra.Command{
		Use:   "supergloo",
		Short: "manage mesh resources with supergloo",
		Long: `supergloo configures resources used by Supergloo server.
	Find more information at https://solo.io`,
		Version: version,
	}

	pFlags := app.PersistentFlags()
	pFlags.BoolVarP(&opts.Top.Static, "static", "s", false, "disable interactive mode")
	pFlags.StringVarP(&opts.Top.File, "filename", "f", "", "file input")

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		install.Cmd(&opts),
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
