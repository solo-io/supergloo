package gloo

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/installutils/kubeinstall"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type installSyncer struct {
	glooInstaller Installer
	ingressClient v1.MeshIngressClient
	reporter      reporter.Reporter
}

// calling this function with nil is valid and expected outside of tests
func NewInstallSyncer(kubeInstaller kubeinstall.Installer, ingressClient v1.MeshIngressClient, reporter reporter.Reporter) v1.InstallSyncer {
	return &installSyncer{
		glooInstaller: newGlooInstaller(kubeInstaller),
		ingressClient: ingressClient,
		reporter:      reporter,
	}
}

func (s *installSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("gloo-install-syncer-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Stringer())
	defer logger.Infof("end sync %v", snap.Stringer())
	resourceErrs := make(reporter.ResourceErrors)

	installs := snap.Installs.List()
	meshes := snap.Meshes.List()
	meshIngresses := snap.Meshingresses.List()

	// split installs by which are active, inactive (istio only)
	// if more than 1 active install, they get errored
	var enabledInstalls, disabledInstalls v1.InstallList
	for _, install := range installs {
		if _, isIngress := install.InstallType.(*v1.Install_Ingress); isIngress {
			if install.Disabled {
				disabledInstalls = append(disabledInstalls, install)
			} else {
				enabledInstalls = append(enabledInstalls, install)
			}
		}
	}
	// Handle mesh installs
	s.handleDisabledInstalls(ctx, disabledInstalls, resourceErrs, meshes, meshIngresses)
	s.handleActiveInstalls(ctx, enabledInstalls, resourceErrs, meshes, meshIngresses)

	// Handle ingress installs

	logger.Infof("writing reports")
	if err := resourceErrs.Validate(); err != nil {
		logger.Warnf("install sync failed with validation errors: %v", err)
	} else {
		logger.Infof("install sync successful")
	}

	// reporter should handle updates to the installs that happened during ensure
	return s.reporter.WriteReports(ctx, resourceErrs, nil)
}

func (s *installSyncer) handleDisabledInstalls(ctx context.Context,
	disabledInstalls v1.InstallList,
	resourceErrs reporter.ResourceErrors, meshes v1.MeshList, meshIngresses v1.MeshIngressList) {
	logger := contextutils.LoggerFrom(ctx)

	for _, in := range disabledInstalls {
		switch installType := in.InstallType.(type) {
		case *v1.Install_Ingress:
			if installType.Ingress.InstalledIngress == nil {
				// mesh was never installed
				resourceErrs.Accept(in)
				continue
			}
			installedIngress := *installType.Ingress.InstalledIngress
			logger.Infof("ensuring install %v is disabled", in.Metadata.Ref())
			if _, err := s.glooInstaller.EnsureGlooInstall(ctx, in, meshes, meshIngresses); err != nil {
				resourceErrs.AddError(in, errors.Wrapf(err, "uninstall failed"))
			} else {
				resourceErrs.Accept(in)
				if err := s.ingressClient.Delete(installedIngress.Namespace,
					installedIngress.Name,
					clients.DeleteOpts{Ctx: ctx}); err != nil {
					logger.Errorf("deleting mesh object %v failed after successful uninstall", installedIngress)
				}
			}
		}
	}
}

func (s *installSyncer) handleActiveInstalls(ctx context.Context,
	enabledInstalls v1.InstallList,
	resourceErrs reporter.ResourceErrors, meshes v1.MeshList, meshIngresses v1.MeshIngressList) {
	logger := contextutils.LoggerFrom(ctx)
	var (
		createdIngress *v1.MeshIngress
		activeInstall  *v1.Install
	)
	switch {
	case len(enabledInstalls) == 1:
		in := enabledInstalls[0]
		contextutils.LoggerFrom(ctx).Infof("ensuring install %v is enabled", in.Metadata.Ref())
		meshIngress, err := s.glooInstaller.EnsureGlooInstall(ctx, in, meshes, meshIngresses)
		if err != nil {
			resourceErrs.AddError(in, errors.Wrapf(err, "install failed"))
			return
		}
		resourceErrs.Accept(in)
		createdIngress = meshIngress
		activeInstall = in
	case len(enabledInstalls) > 1:
		for _, in := range enabledInstalls {
			resourceErrs.AddError(in, errors.Errorf("multiple gloo ingress installations "+
				"are not currently supported. active installs: %v", enabledInstalls.NamespacesDotNames()))
		}
	}

	if createdIngress != nil {
		// update resource version if this is an overwrite
		if existingMeshIngress, err := s.ingressClient.Read(createdIngress.Metadata.Namespace,
			createdIngress.Metadata.Name,
			clients.ReadOpts{Ctx: ctx}); err == nil {

			logger.Infof("overwriting previous mesh ingress %v", existingMeshIngress.Metadata.Ref())
			createdIngress.Metadata.ResourceVersion = existingMeshIngress.Metadata.ResourceVersion
		}

		logger.Infof("writing installed ingress %v", createdIngress.Metadata.Ref())
		if _, err := s.ingressClient.Write(createdIngress,
			clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {
			err := errors.Wrapf(err, "writing installed mesh-ingress object %v failed "+
				"after successful install", createdIngress.Metadata.Ref())
			resourceErrs.AddError(activeInstall, err)
			logger.Errorf("%v", err)
		}
	}

}
