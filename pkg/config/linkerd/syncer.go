package linkerd

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/solo-io/supergloo/pkg/translator/linkerd"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type linkerdConfigSyncer struct {
	translator  linkerd.Translator
	reconcilers Reconcilers
	reporter    reporter.Reporter
}

func NewLinkerdConfigSyncer(translator linkerd.Translator, reconcilers Reconcilers, reporter reporter.Reporter) v1.ConfigSyncer {
	return &linkerdConfigSyncer{translator: translator, reconcilers: reconcilers, reporter: reporter}
}

func (s *linkerdConfigSyncer) Sync(ctx context.Context, snap *v1.ConfigSnapshot) error {

	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("linkerd-config-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("routing_rules", len(snap.Routingrules.List())),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	meshConfigs, resourceErrs, err := s.translator.Translate(ctx, snap)
	if err != nil {
		return errors.Wrapf(err, "translation failed")
	}

	if err := resourceErrs.Validate(); err != nil {
		logger.Errorf("invalid user config or internal error: %v", err)
	}

	// we don't need to return here; if the error was related to the mesh, it shouldn't have been
	// added to the meshConfigs. all meshConfigs are considered to be valid

	for mesh, config := range meshConfigs {
		_, ok := mesh.MeshType.(*v1.Mesh_Linkerd)
		if !ok {
			return errors.Errorf("internal error: a non linkerd-mesh appeared in the mesh config snapshot")
		}

		logger.Infof("reconciling config for mesh %v: ", mesh.Metadata.Ref())
		if err := s.reconcilers.ReconcileAll(ctx, config); err != nil {
			return errors.Wrapf(err, "reconciling config for %v", mesh.Metadata.Ref())
		}
	}

	// finally, write reports
	if err := s.reporter.WriteReports(ctx, resourceErrs, nil); err != nil {
		return errors.Wrapf(err, "writing reports")
	}

	logger.Infof("sync completed successfully!")
	return nil
}
