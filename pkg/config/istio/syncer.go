package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	"github.com/solo-io/supergloo/pkg/config/utils"
	"github.com/solo-io/supergloo/pkg/translator/istio"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type istioConfigSyncer struct {
	translator  istio.Translator
	reconcilers Reconcilers
	reporter    reporter.Reporter
}

func NewIstioConfigSyncer(translator istio.Translator, reconcilers Reconcilers, reporter reporter.Reporter) v1.ConfigSyncer {
	return &istioConfigSyncer{translator: translator, reconcilers: reconcilers, reporter: reporter}
}

// crds required for to sync Istio
var RequiredCrds = []string{
	v1alpha1.MeshPolicyCrd.FullName(),
	v1alpha3.DestinationRuleCrd.FullName(),
	v1alpha3.VirtualServiceCrd.FullName(),
	rbacv1alpha1.RbacConfigCrd.FullName(),
	rbacv1alpha1.ServiceRoleCrd.FullName(),
	rbacv1alpha1.ServiceRoleBindingCrd.FullName(),
}

func (s *istioConfigSyncer) ShouldSync(ctx context.Context, _, snap *v1.ConfigSnapshot) bool {
	return utils.ShouldSync(ctx, "istio", RequiredCrds, snap.Customresourcedefinition)
}

func (s *istioConfigSyncer) Sync(ctx context.Context, snap *v1.ConfigSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translation-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Hash())
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("full snapshot: %v", snap)

	meshConfigs, resourceErrs, err := s.translator.Translate(ctx, snap)
	if err != nil {
		return errors.Wrapf(err, "translation failed")
	}

	// we don't need to return here; if the error was related to the mesh, it shouldn't have been
	// added to the meshConfigs. all meshConfigs are considered to be valid
	if err := resourceErrs.Validate(); err != nil {
		logger.Errorf("invalid user config or internal error: %v", err)
	}

	// ensure that all resources that may have been created by this syncer are removed
	if len(meshConfigs) == 0 {
		if err := s.reconcilers.ReconcileAll(ctx, &istio.MeshConfig{}); err != nil {
			return err
		}
	} else {
		for mesh, config := range meshConfigs {
			_, ok := mesh.MeshType.(*v1.Mesh_Istio)
			if !ok {
				return errors.Errorf("internal error: a non istio-mesh appeared in the mesh config snapshot")
			}

			logger.Infof("reconciling config for mesh: %v", mesh.Metadata.Ref())
			if err := s.reconcilers.ReconcileAll(ctx, config); err != nil {
				logger.Errorf("failed to reconcile config for mesh %s: %v", mesh.Metadata.Ref().Key(), err)
				resourceErrs.AddError(mesh, errors.Wrapf(err, "reconciling configuration"))
			}
		}
	}

	// finally, write reports
	if err := s.reporter.WriteReports(ctx, resourceErrs, nil); err != nil {
		return errors.Wrapf(err, "writing reports")
	}

	logger.Infof("sync completed successfully!")
	return nil
}
