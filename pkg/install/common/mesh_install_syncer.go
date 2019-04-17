package common

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"go.uber.org/zap"
)

type MeshInstallSyncer struct {
	name              string
	meshClient        v1.MeshClient
	reporter          reporter.Reporter
	isOurInstallType  IsInstallType
	ensureMeshInstall EnsureMeshInstall
}

func NewMeshInstallSyncer(name string, meshClient v1.MeshClient, reporter reporter.Reporter, isOurInstallType IsInstallType, ensureMeshInstall EnsureMeshInstall) *MeshInstallSyncer {
	return &MeshInstallSyncer{name: name, meshClient: meshClient, reporter: reporter, isOurInstallType: isOurInstallType, ensureMeshInstall: ensureMeshInstall}
}

type IsInstallType func(install *v1.Install) bool
type EnsureMeshInstall func(ctx context.Context, install *v1.Install, meshes v1.MeshList) (*v1.Mesh, error)

func (s *MeshInstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("%v-install-syncer-%v", s.name, snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("ingresses", len(snap.Meshingresses.List())),
		zap.Int("installs", len(snap.Installs.List())),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
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
		// TODO(EItanya): fix how this works, currently it doesn't reconcile the resources properly
		// Therefore this link isn't maintained across syncs
		addMeshToInstall(install, meshes)
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

	s.handleActiveInstalls(ctx, enabledInstalls, meshes, resourceErrs)

	// reconcile install resources

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

func addMeshToInstall(in *v1.Install, meshes v1.MeshList) {
	if meshInstall := in.GetMesh(); meshInstall != nil {
		if meshInstall.InstalledMesh == nil {
			for _, mesh := range meshes {
				switch meshType := mesh.GetMeshType().(type) {
				case *v1.Mesh_Istio:
					if meshType.Istio.GetInstallationNamespace() == in.GetInstallationNamespace() &&
						mesh.Metadata.Name == in.Metadata.Name {
						meshInstall.InstalledMesh = &core.ResourceRef{
							Name:      mesh.Metadata.Name,
							Namespace: mesh.Metadata.Namespace,
						}
					}
				case *v1.Mesh_LinkerdMesh:
					if meshType.LinkerdMesh.GetInstallationNamespace() == in.GetInstallationNamespace() &&
						mesh.Metadata.Name == in.Metadata.Name {
						meshInstall.InstalledMesh = &core.ResourceRef{
							Name:      mesh.Metadata.Name,
							Namespace: mesh.Metadata.Namespace,
						}
					}
				}
			}
		}
	}
}
