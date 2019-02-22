package install

import (
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
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "install a service mesh using Supergloo",
		Long: `Creates an Install resource which the supergloo controller 
will use to install a service mesh.

Installs represent a desired installation of a supported mesh.
Supergloo watches for installs and synchronizes the managed installations
with the desired configuration in the install object.

Updating the configuration of an install object will cause supergloo to 
modify the corresponding mesh.
`,
	}

	cmd.AddCommand(installIstioSubcommand(opts))
	return cmd
}

func installIstioSubcommand(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "istio",
		Short: "install the Istio control plane",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata(&opts.Create.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyIstioInstall(&opts.Create.InputInstall); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createInstall(opts)
		},
	}
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	return cmd
}

func createInstall(opts *options.Options) error {
	in, err := installFromOpts(opts)
	if err != nil {
		return err
	}
	in, err = helpers.MustInstallClient().Write(in, clients.WriteOpts{})
	if err != nil {
		return err
	}

	helpers.PrintInstalls(v1.InstallList{in}, opts.OutputType)

	return nil
}

func installFromOpts(opts *options.Options) (*v1.Install, error) {
	in := &v1.Install{
		Metadata: opts.Create.Metadata,
		InstallType: &v1.Install_Istio_{
			Istio: &opts.Create.InputInstall.IstioInstall,
		},
	}

	return in, nil
}
