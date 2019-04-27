package common

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
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

type IsInstallType func(mesh *v1.Mesh) *v1.InstallOptions
type EnsureMeshInstall func(ctx context.Context, mesh *v1.Mesh) error

func (s *MeshInstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("%v-install-syncer-%v", s.name, snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	resourceErrs := make(reporter.ResourceErrors)

	meshes := snap.Meshes.List()

	// split installs by which are active, inactive
	// if more than 1 active install of a single mesh type, they get errored
	var enabledInstalls, disabledInstalls v1.MeshList
	for _, mesh := range meshes {
		if options := s.isOurInstallType(mesh); options != nil {
			if options.Disabled {
				disabledInstalls = append(disabledInstalls, mesh)
			} else {
				enabledInstalls = append(enabledInstalls, mesh)
			}
		}
	}
	// perform uninstalls first
	for _, in := range disabledInstalls {
		logger.Infof("ensuring install %v is disabled", in.Metadata.Ref())
		if err := s.ensureMeshInstall(ctx, in); err != nil {
			resourceErrs.AddError(in, errors.Wrapf(err, "uninstall failed"))
		} else {
			resourceErrs.Accept(in)
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
	enabledInstalls v1.MeshList,
	meshes v1.MeshList,
	resourceErrs reporter.ResourceErrors) {

	switch {
	case len(enabledInstalls) == 1:
		in := enabledInstalls[0]
		contextutils.LoggerFrom(ctx).Infof("ensuring install %v is enabled", in.Metadata.Ref())
		err := s.ensureMeshInstall(ctx, in)
		if err != nil {
			resourceErrs.AddError(in, errors.Wrapf(err, "install failed"))
			return
		}
		resourceErrs.Accept(in)

		return
	case len(enabledInstalls) > 1:
		for _, in := range enabledInstalls {
			resourceErrs.AddError(in, errors.Errorf("multiple active %v installations "+
				"are not currently supported. active installs: %v", s.name, enabledInstalls.NamespacesDotNames()))
		}
	}
}
