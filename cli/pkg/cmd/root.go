package cmd

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/cmd/apply"
	"github.com/solo-io/supergloo/cli/pkg/cmd/create"
	"github.com/solo-io/supergloo/cli/pkg/cmd/initialize"
	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
	"github.com/solo-io/supergloo/cli/pkg/cmd/set"
	"github.com/solo-io/supergloo/cli/pkg/cmd/uninstall"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func SuperglooCli(version string) *cobra.Command {
	opts := &options.Options{
		Ctx: context.Background(),
	}

	app := &cobra.Command{
		Use:   "supergloo",
		Short: "CLI for Supergloo",
		Long: `supergloo configures resources watched by the Supergloo Controller.
	Find more information at https://solo.io`,
		Version: version,
	}

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		initialize.Cmd(opts),
		install.Cmd(opts),
		uninstall.Cmd(opts),
		apply.Cmd(opts),
		create.Cmd(opts),
		set.Cmd(opts),
	)

	return app
}
