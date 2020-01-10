package common

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/hashutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"go.uber.org/zap"
)

const (
	LinkerdMeshID = "linkerd"
	IstioMeshID   = "istio"
	AppmeshMeshID = "appmesh"
	ConsulMeshID  = "consul"
)

// Can discover a mesh by looking for in-cluster features that are specific to that mesh
type MeshDiscoveryPlugin interface {
	// human-readable type of the mesh we're discovering (e.g., "consul")
	MeshType() string

	// labels to affix to the mesh object that gets written to the cluster
	DiscoveryLabels() map[string]string

	// return a list of mesh objects that, at least:
	// 	- are marked as belonging to the writeNamespace
	//  - do not include proxy sidecar upstreams
	//  - have an entrypoint set
	DesiredMeshes(ctx context.Context, writeNamespace string, snap *v1.DiscoverySnapshot) (v1.MeshList, error)
}

type meshDiscoverySyncer struct {
	writeNamespace        string
	meshReconciler        v1.MeshReconciler
	meshIngressReconciler v1.MeshIngressReconciler
	plugin                MeshDiscoveryPlugin

	lastDesired v1.MeshList
}

func NewDiscoverySyncer(writeNamespace string, meshReconciler v1.MeshReconciler, meshIngressReconciler v1.MeshIngressReconciler,
	plugin MeshDiscoveryPlugin) v1.DiscoverySyncer {

	return &meshDiscoverySyncer{
		writeNamespace:        writeNamespace,
		meshReconciler:        meshReconciler,
		plugin:                plugin,
		meshIngressReconciler: meshIngressReconciler,
	}
}

func (s *meshDiscoverySyncer) ShouldSync(_, new *v1.DiscoverySnapshot) bool {
	// silence any logs ShouldSync might produce
	silentCtx := contextutils.SilenceLogger(context.TODO())

	desired, err := s.desiredMeshes(silentCtx, new)
	if err != nil {
		return true
	}
	return hashutils.HashAll(desired) != hashutils.HashAll(s.lastDesired)
}

func (s *meshDiscoverySyncer) Sync(ctx context.Context, snap *v1.DiscoverySnapshot) error {
	snapHash := hashutils.MustHash(snap)
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("%v-mesh-discovery-%v", s.plugin.MeshType(), snapHash))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("begin sync",
		zap.Int("Upstreams", len(snap.Upstreams)),
		zap.Int("Deployments", len(snap.Deployments)),
		zap.Int("Tlssecrets", len(snap.Tlssecrets)),
		zap.Int("Pods", len(snap.Pods)),
	)
	defer logger.Infow("end sync")
	logger.Debugf("full snapshot: %v", snap)

	desiredMeshes, err := s.desiredMeshes(ctx, snap)
	if err != nil {
		return err
	}

	meshIngresses := v1.MeshIngressList{}

	desiredMeshes.Each(func(mesh *v1.Mesh) {
		meshIngresses = append(meshIngresses, mesh.GetMeshIngress())
	})

	// reconcile meshes
	if err := s.meshReconciler.Reconcile(
		s.writeNamespace,
		desiredMeshes,
		reconcileMeshDiscoveryMetadata,
		clients.ListOpts{Ctx: ctx, Selector: s.plugin.DiscoveryLabels()},
	); err != nil {
		return err
	}

	// reconcile ingresses
	if err := s.meshIngressReconciler.Reconcile(
		s.writeNamespace,
		meshIngresses,
		nil,
		clients.ListOpts{Ctx: ctx, Selector: s.plugin.DiscoveryLabels()},
	); err != nil {
		return err
	}

	s.lastDesired = desiredMeshes

	return nil
}

func (s *meshDiscoverySyncer) desiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	meshes, err := s.plugin.DesiredMeshes(ctx, s.writeNamespace, snap)
	if err != nil {
		return nil, err
	}
	return meshes.Sort(), nil
}

// after the first write, i.e. on any update
// we are only interested in updating discovery metadata at this point
// do not want to overwrite user-modified config settings
// smi is also included here as it's not intended to be user-writeable
// a return value of false indicates that, for the purposes of this syncer,
// the resource in storage matches the desired resource
func reconcileMeshDiscoveryMetadata(original, desired *v1.Mesh) (b bool, e error) {
	if desired.DiscoveryMetadata.Equal(original.DiscoveryMetadata) && desired.SmiEnabled == original.SmiEnabled {
		return false, nil
	}
	original.DiscoveryMetadata = desired.DiscoveryMetadata
	original.SmiEnabled = desired.SmiEnabled
	*desired = *original
	return true, nil
}
