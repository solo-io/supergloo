package gloo

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"go.uber.org/zap"
)

type glooMtlsSyncer struct {
	cs       *clientset.Clientset
	reporter reporter.Reporter
	plugins  GlooIngressPlugins
}

func NewGlooRegistrationSyncer(cs *clientset.Clientset, plugins ...GlooIngressPlugin) v1.RegistrationSyncer {
	glooReporter := reporter.NewReporter("gloo-registration-reporter",
		cs.Supergloo.Mesh.BaseClient(),
		cs.Supergloo.MeshIngress.BaseClient(),
	)
	return &glooMtlsSyncer{reporter: glooReporter, cs: cs, plugins: plugins}
}

func (s *glooMtlsSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("gloo-registration-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes)),
		zap.Int("mesh-ingresses", len(snap.Meshingresses)),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)

	var glooMeshIngresses v1.MeshIngressList
	for _, meshIngress := range snap.Meshingresses {
		if _, ok := meshIngress.MeshIngressType.(*v1.MeshIngress_Gloo); ok {
			glooMeshIngresses = append(glooMeshIngresses, meshIngress)
		}
	}

	errs := reporter.ResourceErrors{}
	for _, glooIngress := range glooMeshIngresses {
		if err := s.handleGlooMeshIngress(ctx, glooIngress, snap.Meshes); err != nil {
			errs.AddError(glooIngress, err)
			logger.Errorf("unable to update gloo ingress %v, %s", glooIngress.Metadata, err)
		}
	}

	logger.Infof("sync completed successfully!")
	return s.reporter.WriteReports(ctx, errs, nil)
}

func (s *glooMtlsSyncer) handleGlooMeshIngress(ctx context.Context, meshIngress *v1.MeshIngress, meshes v1.MeshList) error {
	var targetMeshes v1.MeshList
	for _, targetMesh := range meshIngress.Meshes {
		mesh, err := meshes.Find(targetMesh.GetNamespace(), targetMesh.GetName())
		if err != nil {
			return errors.Wrapf(err, "could not find mesh %v", targetMesh.Key())
		}
		targetMeshes = append(targetMeshes, mesh)
	}

	for _, plugin := range s.plugins {
		if err := plugin.HandleMeshes(ctx, meshIngress, targetMeshes); err != nil {
			return err
		}
	}
	return nil
}
