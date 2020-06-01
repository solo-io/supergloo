package vm_validation

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/compute-target/snapshot"
	"go.uber.org/zap"
)

var (
	MeshTypeNotSupportedError = func(meshName string) error {
		return eris.Errorf("Mesh type for virtual mesh %s is not currently supported.", meshName)
	}
)

func NewVirtualMeshValidator(
	meshFinder VirtualMeshFinder,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
) snapshot.MeshNetworkingSnapshotValidator {
	return &virtualMeshValidator{
		meshFinder:        meshFinder,
		virtualMeshClient: virtualMeshClient,
	}
}

type virtualMeshValidator struct {
	meshFinder        VirtualMeshFinder
	virtualMeshClient zephyr_networking.VirtualMeshClient
}

func (m *virtualMeshValidator) ValidateVirtualMeshUpsert(ctx context.Context, obj *zephyr_networking.VirtualMesh, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	err := m.validate(ctx, obj)
	m.updateVirtualMeshStatus(ctx, obj)
	return err == nil
}

func (m *virtualMeshValidator) ValidateVirtualMeshDelete(ctx context.Context, obj *zephyr_networking.VirtualMesh, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	err := m.validate(ctx, obj)
	m.updateVirtualMeshStatus(ctx, obj)
	return err == nil
}

func (m *virtualMeshValidator) ValidateMeshServiceUpsert(ctx context.Context, obj *zephyr_discovery.MeshService, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) ValidateMeshServiceDelete(ctx context.Context, obj *zephyr_discovery.MeshService, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) ValidateMeshWorkloadUpsert(ctx context.Context, obj *zephyr_discovery.MeshWorkload, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) ValidateMeshWorkloadDelete(ctx context.Context, obj *zephyr_discovery.MeshWorkload, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) validate(ctx context.Context, vm *zephyr_networking.VirtualMesh) error {
	// TODO: Currently we are listing meshes from all namespaces, however, the namespace(s) should be configurable.
	matchingMeshes, err := m.meshFinder.GetMeshesForVirtualMesh(ctx, vm)
	if err != nil {
		vm.Status.ConfigStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_INVALID,
			Message: err.Error(),
		}
		return err
	}
	for _, v := range matchingMeshes {
		if v.Spec.GetIstio() == nil && v.Spec.GetAwsAppMesh() == nil {
			wrapped := MeshTypeNotSupportedError(v.GetName())
			vm.Status.ConfigStatus = &zephyr_core_types.Status{
				State:   zephyr_core_types.Status_INVALID,
				Message: wrapped.Error(),
			}
			return wrapped
		}
	}

	vm.Status.ConfigStatus = &zephyr_core_types.Status{
		State: zephyr_core_types.Status_ACCEPTED,
	}

	return nil
}

// once the virtual mesh has had its config status updated, call this function to write it into the cluster
func (m *virtualMeshValidator) updateVirtualMeshStatus(ctx context.Context, virtualMesh *zephyr_networking.VirtualMesh) {
	logger := contextutils.LoggerFrom(ctx)

	err := m.virtualMeshClient.UpdateVirtualMeshStatus(ctx, virtualMesh)
	if err != nil {
		logger.Errorw("Error updating validation status on virtual mesh",
			zap.Any("object_meta", virtualMesh.ObjectMeta),
			zap.Any("status", virtualMesh.Status),
		)
	}
}
