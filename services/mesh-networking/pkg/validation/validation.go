package group_validation

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
)

var (
	OnlyIstioSupportedError = func(meshName string) error {
		return eris.Errorf("Illegal mesh type found for group %s, currently only Istio is supported", meshName)
	}
)

func NewMeshGroupValidator(
	meshFinder GroupMeshFinder,
	meshGroupClient zephyr_networking.MeshGroupClient,
) snapshot.MeshNetworkingSnapshotValidator {
	return &meshGroupValidator{
		meshFinder:      meshFinder,
		meshGroupClient: meshGroupClient,
	}
}

type meshGroupValidator struct {
	meshFinder      GroupMeshFinder
	meshGroupClient zephyr_networking.MeshGroupClient
}

func (m *meshGroupValidator) ValidateMeshGroupUpsert(ctx context.Context, obj *networking_v1alpha1.MeshGroup, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	if err := m.validate(ctx, obj); err != nil {
		m.updateMeshGroupStatus(ctx, obj)
		return false
	}
	return true
}

func (m *meshGroupValidator) ValidateMeshGroupDelete(ctx context.Context, obj *networking_v1alpha1.MeshGroup, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	if err := m.validate(ctx, obj); err != nil {
		m.updateMeshGroupStatus(ctx, obj)
		return false
	}
	return true
}

func (m *meshGroupValidator) ValidateMeshServiceUpsert(ctx context.Context, obj *discovery_v1alpha1.MeshService, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *meshGroupValidator) ValidateMeshServiceDelete(ctx context.Context, obj *discovery_v1alpha1.MeshService, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *meshGroupValidator) ValidateMeshWorkloadUpsert(ctx context.Context, obj *discovery_v1alpha1.MeshWorkload, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *meshGroupValidator) ValidateMeshWorkloadDelete(ctx context.Context, obj *discovery_v1alpha1.MeshWorkload, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *meshGroupValidator) validate(ctx context.Context, mg *networking_v1alpha1.MeshGroup) error {
	// TODO: Currently we are listing meshes from all namespaces, however, the namespace(s) should be configurable.
	matchingMeshes, err := m.meshFinder.GetMeshesForGroup(ctx, mg)
	if err != nil {
		mg.Status.ConfigStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: err.Error(),
		}
		return err
	}
	for _, v := range matchingMeshes {
		if v.Spec.GetIstio() == nil {
			wrapped := OnlyIstioSupportedError(v.GetName())
			mg.Status.ConfigStatus = &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_INVALID,
				Message: wrapped.Error(),
			}
			return wrapped
		}
	}
	return nil
}

// once the mesh group has had its config status updated, call this function to write it into the cluster
func (m *meshGroupValidator) updateMeshGroupStatus(ctx context.Context, meshGroup *networking_v1alpha1.MeshGroup) {
	logger := contextutils.LoggerFrom(ctx)

	err := m.meshGroupClient.UpdateStatus(ctx, meshGroup)
	if err != nil {
		logger.Errorf("Error updating validation status on mesh group %+v", meshGroup.ObjectMeta)
	}
}
