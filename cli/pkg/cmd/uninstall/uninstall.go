package uninstall

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uninstall",
		Aliases: []string{"u"},
		Short:   "uninstall a service mesh using Supergloo",
		Long: `Disables an Install resource which the Supergloo controller 
will use to uninstall an installed service mesh.

This only works for meshes that Supergloo installed.

Installs represent a desired installation of a supported mesh.
Supergloo watches for installs and synchronizes the managed installations
with the desired configuration in the install object. When an install is 
disabled, Supergloo will remove corresponding installed components from the cluster.
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyUninstall(opts); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return disableInstall(opts)
		},
	}

	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Uninstall.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	return cmd
}

func disableInstall(opts *options.Options) error {
	in, err := disableInstallFromOpts(opts)
	if err != nil {
		return err
	}
	in, err = helpers.MustInstallClient().Write(in, clients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}

	helpers.PrintInstalls(v1.InstallList{in}, opts.OutputType)

	return nil
}

func disableInstallFromOpts(opts *options.Options) (*v1.Install, error) {
	installToDisable, err := helpers.MustInstallClient().Read(opts.Uninstall.Metadata.Namespace,
		opts.Uninstall.Metadata.Name,
		clients.ReadOpts{Ctx: opts.Ctx})
	if err != nil {
		return nil, err
	}
	if installToDisable.Disabled {
		return nil, errors.Errorf("install %v is already disabled", opts.Uninstall.Metadata)
	}
	installToDisable.Disabled = true

	return installToDisable, nil
}
