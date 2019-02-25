package istio

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type installSyncer struct {
	istioInstaller Installer
	meshClient     v1.MeshClient
	reptorter      reporter.Reporter
}

// calling this function with nil is valid and expected outside of tests
func NewInstallSyncer(istioInstaller Installer, meshClient v1.MeshClient, reptorter reporter.Reporter) v1.InstallSyncer {
	if istioInstaller == nil {
		istioInstaller = &defaultIstioInstaller{}
	}
	return &installSyncer{
		istioInstaller: istioInstaller,
		meshClient:     meshClient,
		reptorter:      reptorter,
	}
}

func (s *installSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "istio-install-syncer")
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Stringer())
	defer logger.Infof("end sync %v", snap.Stringer())
	resourceErrs := make(reporter.ResourceErrors)

	installs := snap.Installs.List()

	// split installs by which are active, inactive (istio only)
	// if more than 1 active install, they get errored
	var enabledInstalls, disabledInstalls v1.InstallList
	for _, install := range installs {
		_, isIstio := install.InstallType.(*v1.Install_Istio_)
		if isIstio {
			if install.Disabled {
				disabledInstalls = append(disabledInstalls, install)
			} else {
				enabledInstalls = append(enabledInstalls, install)
			}
		}
	}

	// perform uninstalls first
	for _, in := range disabledInstalls {
		if in.InstalledMesh == nil {
			// mesh was never installed
			continue
		}
		installedMesh := *in.InstalledMesh
		logger.Infof("ensuring install %v is disabled", in.Metadata.Ref())
		if _, err := s.istioInstaller.EnsureIstioInstall(ctx, in); err != nil {
			resourceErrs.AddError(in, err)
		} else {
			resourceErrs.Accept(in)
			if err := s.meshClient.Delete(installedMesh.Namespace,
				installedMesh.Name,
				clients.DeleteOpts{Ctx: ctx}); err != nil {
				logger.Errorf("deleting mesh objet %v failed after successful uninstall", installedMesh)
			}
		}
	}

	createdMesh := s.handleActiveInstalls(ctx, enabledInstalls, resourceErrs)

	if createdMesh != nil {
		// update resource version if this is an overwrite
		if existingMesh, err := s.meshClient.Read(createdMesh.Metadata.Namespace,
			createdMesh.Metadata.Name,
			clients.ReadOpts{Ctx: ctx}); err == nil {

			logger.Infof("overwriting previous mesh %v", existingMesh.Metadata)
			createdMesh.Metadata.ResourceVersion = existingMesh.Metadata.ResourceVersion

		}

		logger.Infof("writing installed mesh %v", createdMesh.Metadata)
		if _, err := s.meshClient.Write(createdMesh,
			clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {

		}
	}

	logger.Infof("writing reports")
	if err := resourceErrs.Validate(); err != nil {
		logger.Warnf("install sync failed with validation errors: %v", err)
	} else {
		logger.Infof("install sync successful")
	}

	return s.reptorter.WriteReports(ctx, resourceErrs, nil)

}

func (s *installSyncer) handleActiveInstalls(ctx context.Context,
	enabledInstalls v1.InstallList,
	resourceErrs reporter.ResourceErrors) *v1.Mesh {

	switch {
	case len(enabledInstalls) == 1:
		in := enabledInstalls[0]
		contextutils.LoggerFrom(ctx).Infof("ensuring install %v is enabled", in.Metadata.Ref())
		mesh, err := s.istioInstaller.EnsureIstioInstall(ctx, in)
		if err != nil {
			resourceErrs.AddError(in, err)
			return nil
		}
		return mesh
	case len(enabledInstalls) > 1:
		for _, in := range enabledInstalls {
			resourceErrs.AddError(in, errors.Errorf("multiple active istio installactions "+
				"are not currently supported. active installs: %v", enabledInstalls.NamespacesDotNames()))
		}
	}
	return nil
}
