package cmd

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/install"
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
		faultInjection(&opts),
		loadBalancing(&opts),
		retries(&opts),
	)

	return app
}

func faultInjection(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fault-injection",
		Short: `stress test your mesh with faults`,
		Long:  `stress test your mesh with faults`,
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("this feature will be available in 2019")
		},
	}
	linkMeshToolFlags(cmd, opts)
	return cmd
}

func loadBalancing(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load-balancing",
		Short: `specify traffic distribution`,
		Long:  `specify traffic distribution`,
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("this feature will be available in 2019")
		},
	}
	linkMeshToolFlags(cmd, opts)
	return cmd
}

func retries(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retries",
		Short: `configure retry parameters`,
		Long:  `configure retry parameters`,
		Run: func(c *cobra.Command, args []string) {
			meshToolPlaceholder(opts)
		},
	}
	linkMeshToolFlags(cmd, opts)
	return cmd
}

func linkMeshToolFlags(cmd *cobra.Command, opts *options.Options) {
	pflags := cmd.PersistentFlags()
	pflags.StringVar(&opts.MeshTool.MeshId, "meshid", "", "mesh to modify")
	pflags.StringVar(&opts.MeshTool.ServiceId, "serviceid", "", "service to modify")
}

func meshToolPlaceholder(opts *options.Options) {
	fmt.Println("this feature will be available in 2019")
}
