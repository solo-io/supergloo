package install

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/enterprise"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/internal/flags"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &flags.Options{}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Gloo Mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(ctx, opts)
		},
	}
	opts.AddToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	cmd.AddCommand(enterprise.Command(ctx))
	return cmd
}

func install(ctx context.Context, opts *flags.Options) error {
	// User-specified chartPath takes precedence over specified version.
	gloomeshChartUri := opts.ChartPath
	gloomeshVersion := opts.Version
	if opts.Version == "" {
		gloomeshVersion = version.Version
	}
	if gloomeshChartUri == "" {
		gloomeshChartUri = fmt.Sprintf(gloomesh.GlooMeshChartUriTemplate, gloomeshVersion)
	}

	if err := opts.GetInstaller(gloomeshChartUri).InstallGlooMesh(ctx); err != nil {
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
