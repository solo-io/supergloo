package enterprise

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/internal/flags"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Install Gloo Mesh enterprise",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Verbose = globalFlags.Verbose
			return install(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	flags.Options
	licenseKey string
	skipUI     bool
	skipRBAC   bool
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	o.AddToFlags(flags)
	flags.StringVar(&o.licenseKey, "license", "", "Gloo Mesh Enterprise license key")
	cobra.MarkFlagRequired(flags, "license")
	flags.BoolVar(&o.skipUI, "skip-ui", false, "Skip installation of the Gloo Mesh UI")
	flags.BoolVar(&o.skipRBAC, "skip-rbac", false, "Skip installation of the RBAC Webhook")
}

func install(ctx context.Context, opts *options) error {
	const (
		repoURI   = "https://storage.googleapis.com/gloo-mesh-enterprise"
		chartName = "gloo-mesh-enterprise"
	)
	if opts.Version == "" {
		version, err := helm.GetLatestChartVersion(repoURI, chartName, true)
		if err != nil {
			return err
		}
		opts.Version = version
	}

	installer := opts.GetInstaller(gloomesh.GlooMeshEnterpriseChartUriTemplate)
	installer.Values["licenseKey"] = opts.licenseKey
	if opts.skipUI {
		installer.Values["gloo-mesh-ui.enabled"] = "false"
	}
	if opts.skipRBAC {
		installer.Values["rbac-webhook.enabled"] = "false"
	}
	if err := installer.InstallGlooMeshEnterprise(ctx); err != nil {
		return eris.Wrap(err, "installing gloo-mesh-enterprise")
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
