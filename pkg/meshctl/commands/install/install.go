package install

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/internal/flags"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &flags.Options{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Gloo Mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Verbose = globalFlags.Verbose
			return install(ctx, opts)
		},
	}
	opts.AddToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	cmd.AddCommand(enterprise.Command(ctx, globalFlags))
	return cmd
}

func install(ctx context.Context, opts *flags.Options) error {
	if opts.Version == "" {
		opts.Version = version.Version
	}
	if err := opts.GetInstaller(gloomesh.GlooMeshChartUriTemplate).InstallGlooMesh(ctx); err != nil {
		return eris.Wrap(err, "installing gloo-mesh")
	}

	if opts.Register && !opts.DryRun {
		registrantOpts := opts.GetRegistrationOptions()
		registrant, err := registration.NewRegistrant(&registrantOpts)
		if err != nil {
			return eris.Wrap(err, "initializing registrant")
		}
		if err := registrant.RegisterCluster(ctx); err != nil {
			return eris.Wrap(err, "registering management-plane cluster")
		}
	}
	return nil
}
