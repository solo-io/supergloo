package selector

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
)

// Key that uniquely identifies a MeshService, used for equality checks between two MeshServices.
type MeshServiceId struct {
	Name        string
	Namespace   string
	ClusterName string
}

func (m *MeshServiceId) Equals(that *MeshServiceId) bool {
	if that == nil {
		return m == nil
	} else if m == nil {
		return false
	}
	return *m == *that
}

// Construct a key uniquely identifying a MeshService, used to check equality between two MeshServices.
func BuildIdForMeshService(
	ctx context.Context,
	meshClient zephyr_discovery.MeshClient,
	meshService *zephyr_discovery.MeshService,
) (*MeshServiceId, error) {
	mesh, err := meshClient.GetMeshMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
	if err != nil {
		return nil, err
	}
	return &MeshServiceId{
		Name:        meshService.GetName(),
		Namespace:   meshService.GetNamespace(),
		ClusterName: mesh.Spec.GetCluster().GetName(),
	}, nil
}
