package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type istioConfigSyncer struct {
	translator  Translator
	reconcilers Reconcilers
	reporter    reporter.Reporter
}

func NewIstioConfigSyncer(translator Translator, reconcilers Reconcilers, reporter reporter.Reporter) v1.ConfigSyncer {
	return &istioConfigSyncer{translator: translator, reconcilers: reconcilers, reporter: reporter}
}

func (s *istioConfigSyncer) Sync(ctx context.Context, snap *v1.ConfigSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translation-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v: %v", snap.Hash(), snap.Stringer())
	defer logger.Infof("end sync %v", snap.Hash())
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
		istio, ok := mesh.MeshType.(*v1.Mesh_Istio)
		if !ok {
			return errors.Errorf("internal error: a non istio-mesh appeared in the mesh config snapshot")
		}

		// NOTE (ilackarms): eventually we want a way to correlate the mesh with the cluster it's installed to
		writeNamespace := istio.Istio.InstallationNamespace

		logger.Infof("reconciling config for mesh %v: ", mesh.Metadata.Ref())
		if err := s.reconcilers.ReconcileAll(ctx, writeNamespace, config); err != nil {
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
