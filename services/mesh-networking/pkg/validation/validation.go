package vm_validation

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster/snapshot"
	"go.uber.org/zap"
)

var (
	OnlyIstioSupportedError = func(meshName string) error {
		return eris.Errorf("Illegal mesh type found for virtual mesh %s, currently only Istio is supported", meshName)
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

func (m *virtualMeshValidator) ValidateVirtualMeshUpsert(ctx context.Context, obj *networking_v1alpha1.VirtualMesh, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	if err := m.validate(ctx, obj); err != nil {
		m.updateVirtualMeshStatus(ctx, obj)
		return false
	}
	return true
}

func (m *virtualMeshValidator) ValidateVirtualMeshDelete(ctx context.Context, obj *networking_v1alpha1.VirtualMesh, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	if err := m.validate(ctx, obj); err != nil {
		m.updateVirtualMeshStatus(ctx, obj)
		return false
	}
	return true
}

func (m *virtualMeshValidator) ValidateMeshServiceUpsert(ctx context.Context, obj *discovery_v1alpha1.MeshService, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) ValidateMeshServiceDelete(ctx context.Context, obj *discovery_v1alpha1.MeshService, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) ValidateMeshWorkloadUpsert(ctx context.Context, obj *discovery_v1alpha1.MeshWorkload, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) ValidateMeshWorkloadDelete(ctx context.Context, obj *discovery_v1alpha1.MeshWorkload, snapshot *snapshot.MeshNetworkingSnapshot) bool {
	return true
}

func (m *virtualMeshValidator) validate(ctx context.Context, vm *networking_v1alpha1.VirtualMesh) error {
	// TODO: Currently we are listing meshes from all namespaces, however, the namespace(s) should be configurable.
	matchingMeshes, err := m.meshFinder.GetMeshesForVirtualMesh(ctx, vm)
	if err != nil {
		vm.Status.ConfigStatus = &core_types.Status{
			State:   core_types.Status_INVALID,
			Message: err.Error(),
		}
		return err
	}
	for _, v := range matchingMeshes {
		if v.Spec.GetIstio() == nil {
			wrapped := OnlyIstioSupportedError(v.GetName())
			vm.Status.ConfigStatus = &core_types.Status{
				State:   core_types.Status_INVALID,
				Message: wrapped.Error(),
			}
			return wrapped
		}
	}
	return nil
}

// once the virtual mesh has had its config status updated, call this function to write it into the cluster
func (m *virtualMeshValidator) updateVirtualMeshStatus(ctx context.Context, virtualMesh *networking_v1alpha1.VirtualMesh) {
	logger := contextutils.LoggerFrom(ctx)

	err := m.virtualMeshClient.UpdateStatus(ctx, virtualMesh)
	if err != nil {
		logger.Errorw("Error updating validation status on virtual mesh",
			zap.Any("object_meta", virtualMesh.ObjectMeta),
			zap.Any("status", virtualMesh.Status),
		)
	}
}
