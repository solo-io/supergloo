package selection

import (
	"context"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
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
	meshClient smh_discovery.MeshClient,
	meshService *smh_discovery.MeshService,
) (*MeshServiceId, error) {
	mesh, err := meshClient.GetMesh(ctx, ResourceRefToObjectKey(meshService.Spec.GetMesh()))
	if err != nil {
		return nil, err
	}
	return &MeshServiceId{
		Name:        meshService.GetName(),
		Namespace:   meshService.GetNamespace(),
		ClusterName: mesh.Spec.GetCluster().GetName(),
	}, nil
}
