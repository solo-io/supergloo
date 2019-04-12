package common

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type MeshInstallSyncer struct {
	name              string
	meshClient        v1.MeshClient
	reporter          reporter.Reporter
	isOurInstallType  func(install *v1.Install) bool
	ensureMeshInstall func(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error)
}

func NewMeshInstallSyncer(name string, meshClient v1.MeshClient, reporter reporter.Reporter, isOurInstallType func(install *v1.Install) bool, ensureMeshInstall func(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error)) *MeshInstallSyncer {
	return &MeshInstallSyncer{name: name, meshClient: meshClient, reporter: reporter, isOurInstallType: isOurInstallType, ensureMeshInstall: ensureMeshInstall}
}

func (s *MeshInstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("%v-install-syncer-%v", s.name, snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Stringer())
	defer logger.Infof("end sync %v", snap.Stringer())
	resourceErrs := make(reporter.ResourceErrors)

	installs := snap.Installs.List()
	meshes := snap.Meshes.List()

	// split installs by which are active, inactive
	// if more than 1 active install of a single mesh type, they get errored
	var enabledInstalls, disabledInstalls v1.InstallList
	for _, install := range installs {
		if install.GetMesh() == nil {
			continue
		}
		if s.isOurInstallType(install) {
			if install.Disabled {
				disabledInstalls = append(disabledInstalls, install)
			} else {
				enabledInstalls = append(enabledInstalls, install)
			}
		}
	}
	// perform uninstalls first
	for _, in := range disabledInstalls {
		installMesh := in.GetMesh()
		if installMesh.InstalledMesh == nil {
			// mesh was never installed
			resourceErrs.Accept(in)
			continue
		}
		installedMesh := *installMesh.InstalledMesh
		logger.Infof("ensuring install %v is disabled", in.Metadata.Ref())
		if _, err := s.ensureMeshInstall(ctx, in, meshes); err != nil {
			resourceErrs.AddError(in, errors.Wrapf(err, "uninstall failed"))
		} else {
			resourceErrs.Accept(in)
			if err := s.meshClient.Delete(installedMesh.Namespace,
				installedMesh.Name,
				clients.DeleteOpts{Ctx: ctx}); err != nil {
				logger.Errorf("deleting mesh object %v failed after successful uninstall", installedMesh)
			}
		}
	}

	createdMesh, activeInstall := s.handleActiveInstalls(ctx, enabledInstalls, meshes, resourceErrs)

	if createdMesh != nil {
		// update resource version if this is an overwrite
		if existingMesh, err := s.meshClient.Read(createdMesh.Metadata.Namespace,
			createdMesh.Metadata.Name,
			clients.ReadOpts{Ctx: ctx}); err == nil {

			logger.Infof("overwriting previous mesh %v", existingMesh.Metadata.Ref())
			createdMesh.Metadata.ResourceVersion = existingMesh.Metadata.ResourceVersion
		}

		logger.Infof("writing installed mesh %v", createdMesh.Metadata.Ref())
		if _, err := s.meshClient.Write(createdMesh,
			clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {
			err := errors.Wrapf(err, "writing installed mesh object %v failed "+
				"after successful install", createdMesh.Metadata.Ref())
			resourceErrs.AddError(activeInstall, err)
			logger.Errorf("%v", err)
		}

		// caller should expect the install to have been modified
		ref := createdMesh.Metadata.Ref()
		activeInstall.GetMesh().InstalledMesh = &ref
	}

	logger.Infof("writing reports")
	if err := resourceErrs.Validate(); err != nil {
		logger.Warnf("install sync failed with validation errors: %v", err)
	} else {
		logger.Infof("install sync successful")
	}

	// reporter should handle updates to the installs that happened during ensure
	return s.reporter.WriteReports(ctx, resourceErrs, nil)

}

func (s *MeshInstallSyncer) handleActiveInstalls(ctx context.Context,
	enabledInstalls v1.InstallList,
	meshes v1.MeshList,
	resourceErrs reporter.ResourceErrors) (*v1.Mesh, *v1.Install) {

	switch {
	case len(enabledInstalls) == 1:
		in := enabledInstalls[0]
		contextutils.LoggerFrom(ctx).Infof("ensuring install %v is enabled", in.Metadata.Ref())
		mesh, err := s.ensureMeshInstall(ctx, in, meshes)
		if err != nil {
			resourceErrs.AddError(in, errors.Wrapf(err, "install failed"))
			return nil, nil
		}
		resourceErrs.Accept(in)

		return mesh, in
	case len(enabledInstalls) > 1:
		for _, in := range enabledInstalls {
			resourceErrs.AddError(in, errors.Errorf("multiple active %v installations "+
				"are not currently supported. active installs: %v", s.name, enabledInstalls.NamespacesDotNames()))
		}
	}
	return nil, nil
}
