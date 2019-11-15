package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/hashutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	zeph_core "github.com/solo-io/mesh-projects/pkg/api/v1/core"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"go.uber.org/zap"
)

const (
	LinkerdMeshID = "linkerd"
	IstioMeshID   = "istio"
	AppmeshMeshID = "appmesh"
)

type MeshDiscoveryPlugin interface {
	MeshType() string
	DiscoveryLabels() map[string]string
	DesiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error)
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
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("%v-mesh-discovery-%v", s.plugin.MeshType(), snap.Hash()))
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
		mesh.Metadata.Namespace = s.writeNamespace
		// remove sidecar upstreams from injection list
		s.removeSidecarUpstreams(mesh)
		meshIngress := s.createMeshIngressForMesh(mesh)
		meshIngresses = append(meshIngresses, meshIngress)
		mesh.EntryPoint = &zeph_core.ClusterResourceRef{
			Resource: meshIngress.Metadata.Ref(),
		}
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

func (s *meshDiscoverySyncer) createMeshIngressForMesh(mesh *v1.Mesh) *v1.MeshIngress {
	return &v1.MeshIngress{
		Metadata: core.Metadata{
			Name:      mesh.Metadata.Name,
			Namespace: mesh.Metadata.Namespace,
			Labels:    s.plugin.DiscoveryLabels(),
		},
		IngressType: &v1.MeshIngress_Gloo{
			Gloo: &v1.MeshIngress_GlooIngress{
				Namespace:   "gloo-system",
				ServiceName: "gateway-proxy-v2",
				Port:        "mesh-bridge",
			},
		},
	}
}

func (s *meshDiscoverySyncer) removeSidecarUpstreams(mesh *v1.Mesh) {
	var removalString string
	switch mesh.GetMeshType().(type) {
	case *v1.Mesh_Istio:
		removalString = IstioMeshID
	case *v1.Mesh_Linkerd:
		removalString = LinkerdMeshID
	case *v1.Mesh_AwsAppMesh:
		removalString = AppmeshMeshID
	}
	removalString = fmt.Sprintf("-%s-", removalString)
	var tempList []*core.ResourceRef
	for _, ref := range mesh.DiscoveryMetadata.Upstreams {
		if !strings.Contains(ref.GetName(), removalString) {
			tempList = append(tempList, ref)
		}
	}
	mesh.DiscoveryMetadata.Upstreams = tempList
}

func (s *meshDiscoverySyncer) desiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	meshes, err := s.plugin.DesiredMeshes(ctx, snap)
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
