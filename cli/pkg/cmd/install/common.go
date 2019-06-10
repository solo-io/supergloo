package install

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/kubeutils"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	apierrs "github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

type postInstallAction func(opts *options.Options) error

func createInstall(opts *options.Options, install *v1.Install, postInstallActions ...postInstallAction) error {

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
		install.InstallationNamespace = existing.InstallationNamespace

		installMeshIngress, installIsMeshIngress := install.InstallType.(*v1.Install_Ingress)

		switch existingType := existing.InstallType.(type) {
		case *v1.Install_Ingress:
			if installIsMeshIngress {
				installMeshIngress.Ingress.InstalledIngress = existingType.Ingress.InstalledIngress
			}
		}

	}
	// create the installation namespace if it does not already exist
	installNamespace := install.InstallationNamespace
	err = kubeutils.CreateNamespacesInParallel(clients.MustKubeClient(), installNamespace)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	install, err = clients.MustInstallClient().Write(install, skclients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}

	// If any post install actions are present, wait for install to be accepted and then perform them in order
	if len(postInstallActions) > 0 {
		timeout := opts.Install.InstallTimeout
		if err := helpers.WaitForInstallStatus(opts.Ctx, install.Metadata.Ref(), core.Status_Accepted, timeout); err != nil {
			return err
		}

		for _, action := range postInstallActions {
			if err := action(opts); err != nil {
				return err
			}
		}
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
