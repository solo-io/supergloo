package install

import (
	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	apierrs "github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/spf13/cobra"
)

func installIstioCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "istio",
		Short: "install the Istio control plane",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata(&opts.Install.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyIstioInstall(&opts.Install.InputInstall); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createInstall(opts)
		},
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Install.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	flagutils.AddIstioInstallFlags(cmd.PersistentFlags(), &opts.Install.InputInstall)
	return cmd
}

func createInstall(opts *options.Options) error {
	// first check if install exists; if so, update and write that object
	in, err := updateDisabledInstall(opts)
	if err != nil {
		return err
	}
	if in == nil {
		in, err = installFromOpts(opts)
		if err != nil {
			return err
		}
	}
	in, err = helpers.MustInstallClient().Write(in, clients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}

	helpers.PrintInstalls(v1.InstallList{in}, opts.OutputType)

	return nil
}

func updateDisabledInstall(opts *options.Options) (*v1.Install, error) {
	existingInstall, err := helpers.MustInstallClient().Read(opts.Install.Metadata.Namespace,
		opts.Install.Metadata.Name, clients.ReadOpts{Ctx: opts.Ctx})
	if err != nil {
		if apierrs.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !existingInstall.Disabled {
		return nil, errors.Errorf("install %v is already installed and enabled", opts.Install.Metadata)
	}
	existingInstall.Disabled = false
	return existingInstall, nil
}

func installFromOpts(opts *options.Options) (*v1.Install, error) {
	if err := validate(opts.Install.InputInstall); err != nil {
		return nil, err
	}
	in := &v1.Install{
		Metadata: opts.Install.Metadata,
		InstallType: &v1.Install_Istio_{
			Istio: &opts.Install.InputInstall.IstioInstall,
		},
	}

	return in, nil
}

func validate(in options.InputInstall) error {
	var validVersion bool
	for _, ver := range []string{
		istio.IstioVersion103,
		istio.IstioVersion105,
	} {
		if in.IstioInstall.IstioVersion == ver {
			validVersion = true
			break
		}
	}
	if !validVersion {
		return errors.Errorf("%v is not a suppported "+
			"istio version", in.IstioInstall.IstioVersion)
	}

	return nil
}
