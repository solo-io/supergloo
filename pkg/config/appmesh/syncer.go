package appmesh

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
)

type appMeshConfigSyncer struct {
	translator appmesh.Translator
	reporter   reporter.Reporter
}

func NewAppMeshConfigSyncer(translator appmesh.Translator, reporter reporter.Reporter) (v1.ConfigSyncer, error) {
	return &appMeshConfigSyncer{translator: translator, reporter: reporter}, nil
}

func (s *appMeshConfigSyncer) Sync(ctx context.Context, snap *v1.ConfigSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("app-mesh-translation-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Hash())
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("full snapshot: %v", snap)

	meshConfigs, resourceErrs, err := s.translator.Translate(ctx, snap)
	if err != nil {
		return errors.Wrapf(err, "AWS App Mesh translation failed")
	}

	if err := resourceErrs.Validate(); err != nil {
		logger.Errorf("invalid configuration or internal error: %v", err)
	}

	// No need to return here; if the error was related to the mesh, it shouldn't have been
	// added to the meshConfigs. all meshConfigs are considered to be valid

	for mesh, config := range meshConfigs {
		if mesh.GetAwsAppMesh() == nil {
			return errors.Errorf("internal error: a non AWS App Mesh appeared in the mesh config snapshot")
		}

		logger.Infof("reconciling config for mesh %v: ", mesh.Metadata.Ref())
		// TODO: implement reconciliation
		logger.Debugf("App Mesh Configuration  is: %v", config)
	}

	// finally, write reports
	if err := s.reporter.WriteReports(ctx, resourceErrs, nil); err != nil {
		return errors.Wrapf(err, "writing reports")
	}

	logger.Infof("sync completed successfully!")
	return nil
}
