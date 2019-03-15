package install

import (
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	apierrs "github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/constants"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func installIstioCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "istio",
		Short: "install the Istio control plane",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("installation", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyIstioInstall(&opts.Install); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createInstall(opts)
		},
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	flagutils.AddIstioInstallFlags(cmd.PersistentFlags(), &opts.Install)
	return cmd
}

func createInstall(opts *options.Options) error {
	install, err := installFromOpts(opts)
	if err != nil {
		return err
	}
	// check for existing install
	// if upgrade is set, upgrade it
	// else error
	existing, err := getExistingInstall(opts)
	if err != nil {
		return err
	}
	if existing != nil {
		if !opts.Install.Update && !existing.Disabled {
			return errors.Errorf("install %v is already installed and enabled", install.Metadata.Ref())
		}
		contextutils.LoggerFrom(opts.Ctx).Infof("upgrading istio install from %s to %s",
			helpers.MustMarshalProto(existing), helpers.MustMarshalProto(install))
		install.Metadata.ResourceVersion = existing.Metadata.ResourceVersion
		install.InstalledManifest = existing.InstalledManifest
		install.InstallationNamespace = existing.InstallationNamespace
		existingMesh, existingIsMesh := existing.InstallType.(*v1.Install_Mesh)
		installMesh, istallIsMesh := install.InstallType.(*v1.Install_Mesh)
		if existingIsMesh && istallIsMesh {
			installMesh.Mesh.InstalledMesh = existingMesh.Mesh.InstalledMesh
		}
	}
	install, err = helpers.MustInstallClient().Write(install, clients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}

	helpers.PrintInstalls(v1.InstallList{install}, opts.OutputType)

	return nil
}

// returns install, nil if install exists
// returns nil, nil if install does not exist
// returns nil, err if other error
func getExistingInstall(opts *options.Options) (*v1.Install, error) {
	existingInstall, err := helpers.MustInstallClient().Read(opts.Metadata.Namespace,
		opts.Metadata.Name, clients.ReadOpts{Ctx: opts.Ctx})
	if err != nil {
		if apierrs.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return existingInstall, nil
}

func installFromOpts(opts *options.Options) (*v1.Install, error) {
	if err := validate(opts.Install); err != nil {
		return nil, err
	}
	in := &v1.Install{
		Metadata: opts.Metadata,
		InstallType: &v1.Install_Mesh{
			Mesh: &v1.MeshInstall{
				InstallType: &v1.MeshInstall_IstioMesh{
					IstioMesh: &opts.Install.IstioInstall,
				},
			},
		},
	}

	return in, nil
}

func validate(in options.Install) error {
	var validVersion bool
	for _, ver := range constants.SupportedIstioVersions {
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
