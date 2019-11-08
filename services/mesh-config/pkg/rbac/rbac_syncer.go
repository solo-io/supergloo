package rbac

import (
	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"go.uber.org/zap"
)

// watch meshes, write mesh (istio only at the moment) RBAC objects

type rbacSyncer struct {
	rbacConfigReconciler v1alpha1.RbacConfigReconciler
	meshReconciler       v1.MeshReconciler
	writeNamespace       string
}

func NewRbacSyncer(writeNamespace string, meshReconciler v1.MeshReconciler, rbacConfigReconciler v1alpha1.RbacConfigReconciler) *rbacSyncer {
	return &rbacSyncer{
		rbacConfigReconciler: rbacConfigReconciler,
		meshReconciler:       meshReconciler,
		writeNamespace:       writeNamespace,
	}
}

func (s *rbacSyncer) Sync(ctx context.Context, snap *v1.RbacSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("rbac-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("begin sync",
		zap.Int("RbacConfigs", len(snap.RbacConfigs)),
		zap.Int("Meshes", len(snap.Meshes)),
	)
	defer logger.Infow("end sync")
	logger.Debugw("full snapshot",
		zap.Any("snap", snap))

	desiredSnap := getDesired(ctx, snap)
	logger.Debugw("desired snapshot",
		zap.Any("Meshes", desiredSnap.Meshes),
		zap.Any("RbacConfigs", desiredSnap.RbacConfigs))
	if err := s.meshReconciler.Reconcile(s.writeNamespace, desiredSnap.Meshes, nil, clients.ListOpts{}); err != nil {
		return err
	}
	if err := s.rbacConfigReconciler.Reconcile("", desiredSnap.RbacConfigs, nil, clients.ListOpts{}); err != nil {
		return err
	}

	return nil
}

func getDesired(ctx context.Context, in *v1.RbacSnapshot) *v1.RbacSnapshot {
	clusterResourceMap := mapRbacConfigByCluster(in)
	out := &v1.RbacSnapshot{}
	for _, m := range in.Meshes {
		switch m.MeshType.(type) {
		case *v1.Mesh_Istio:
			nextMesh, nextRbacConfig := handleIstioRbac(ctx, m, clusterResourceMap[getIstioMeshKeyFromMesh(m)])
			if nextMesh != nil {
				out.Meshes = append(out.Meshes, nextMesh)
			}
			if nextRbacConfig != nil {
				out.RbacConfigs = append(out.RbacConfigs, nextRbacConfig)
			}
		default:
			out.Meshes = append(out.Meshes, handleUnsupportedMeshes(m))
		}
	}
	return out
}

type meshKey string

func getIstioMeshKeyFromMesh(m *v1.Mesh) meshKey {
	return meshKey(fmt.Sprintf("%v-%v", m.DiscoveryMetadata.Cluster, m.Metadata.String()))
}
func getIstioMeshKeyFromRbacConfig(m *v1alpha1.RbacConfig) meshKey {
	return meshKey(fmt.Sprintf("%v-%v", m.Metadata.Cluster, m.Metadata.String()))
}

func mapRbacConfigByCluster(in *v1.RbacSnapshot) map[meshKey]*v1alpha1.RbacConfig {
	// there can only be one RbacConfig per mesh so we do not need to store a list
	m := make(map[meshKey]*v1alpha1.RbacConfig)
	for _, rbacConfig := range in.RbacConfigs {
		rbacConfig := rbacConfig
		m[getIstioMeshKeyFromRbacConfig(rbacConfig)] = rbacConfig
	}
	return m
}
