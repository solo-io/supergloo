package install

import (
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	apierrs "github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func createInstall(opts *options.Options, install *v1.Install) error {

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
		contextutils.LoggerFrom(opts.Ctx).Infof("upgrading install from %s to %s",
			helpers.MustMarshalProto(existing), helpers.MustMarshalProto(install))
		install.Metadata.ResourceVersion = existing.Metadata.ResourceVersion
		install.InstalledManifest = existing.InstalledManifest
		install.InstallationNamespace = existing.InstallationNamespace

		installMesh, installIsMesh := install.InstallType.(*v1.Install_Mesh)
		installMeshIngress, installIsMeshIngress := install.InstallType.(*v1.Install_Ingress)

		switch existingType := existing.InstallType.(type) {
		case *v1.Install_Mesh:
			if installIsMesh {
				installMesh.Mesh.InstalledMesh = existingType.Mesh.InstalledMesh
			}
		case *v1.Install_Ingress:
			if installIsMeshIngress {
				installMeshIngress.Ingress.InstalledIngress = existingType.Ingress.InstalledIngress
			}
		}

	}
	install, err = clients.MustInstallClient().Write(install, skclients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
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
	existingInstall, err := clients.MustInstallClient().Read(opts.Metadata.Namespace,
		opts.Metadata.Name, skclients.ReadOpts{Ctx: opts.Ctx})
	if err != nil {
		if apierrs.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return existingInstall, nil
}
