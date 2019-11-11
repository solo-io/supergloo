package clusterrbac

import (
	"context"
	"fmt"

	"github.com/solo-io/mesh-projects/services/mesh-config/pkg/clusterrbac/translator"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"go.uber.org/zap"
)

// watch meshes, write mesh (istio only at the moment) RBAC objects

type rbacSyncer struct {
	clusterRbacConfigReconciler v1alpha1.ClusterRbacConfigReconciler
	meshReconciler              v1.MeshReconciler
	writeNamespace              string
	translator                  translator.Translator
}

func NewRbacSyncer(writeNamespace string,
	meshReconciler v1.MeshReconciler,
	clusterRbacConfigReconciler v1alpha1.ClusterRbacConfigReconciler,
	translator translator.Translator) *rbacSyncer {
	return &rbacSyncer{
		clusterRbacConfigReconciler: clusterRbacConfigReconciler,
		meshReconciler:              meshReconciler,
		writeNamespace:              writeNamespace,
		translator:                  translator,
	}
}

func (s *rbacSyncer) Sync(ctx context.Context, snap *v1.RbacSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("rbac-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("begin sync",
		zap.Int("Meshes", len(snap.Meshes)),
	)

	defer logger.Infow("end sync")
	logger.Debugw("full snapshot",
		zap.Any("snap", snap))

	desiredClusterRbac, err := s.translator.Translate(ctx, snap)
	if err != nil {
		return err
	}

	if err := s.meshReconciler.Reconcile(s.writeNamespace, snap.Meshes, nil,
		clients.ListOpts{}); err != nil {
		return err
	}
	if err := s.clusterRbacConfigReconciler.Reconcile("", desiredClusterRbac,
		nil, clients.ListOpts{Selector: translator.OwnerLabels}); err != nil {
		return err
	}

	return nil
}
