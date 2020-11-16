package enterprise

import (
	"context"
	"errors"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/internal/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Install Gloo Mesh enterprise",
		RunE: func(cmd *cobra.Command, args []string) error {
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
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	o.AddToFlags(flags)
	flags.StringVar(&o.licenseKey, "license", "", "Enterprise license key")
}

var NoLicenseError = errors.New("Gloo Mesh Enterprise requries a license key.")

func install(ctx context.Context, opts *options) error {
	if opts.licenseKey == "" {
		return NoLicenseError
	}

	return nil
}
