package vm_validation

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster/snapshot"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	OnlyIstioSupportedError = func(meshName string) error {
		return eris.Errorf("Illegal mesh type found for virtual mesh %s, currently only Istio is supported", meshName)
	}
	OnlyHomogenousVirtualMeshesSupported = func(vmName, vmNamespace string, numMeshTypes int) error {
		return eris.Errorf("Virtual mesh %s.%s contains %d different mesh types, but only homogenous virtual meshes (one mesh type) supported", vmName, vmNamespace, numMeshTypes)
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

	representedMeshTypes := sets.NewInt32()
	for _, v := range matchingMeshes {
		if v.Spec.GetIstio1_5() == nil && v.Spec.GetIstio1_6() == nil {
			wrapped := OnlyIstioSupportedError(v.GetName())
			vm.Status.ConfigStatus = &zephyr_core_types.Status{
				State:   zephyr_core_types.Status_INVALID,
				Message: wrapped.Error(),
			}
			return wrapped
		}

		meshType, err := enum_conversion.MeshToMeshType(v)
		if err != nil {
			return err
		}

		representedMeshTypes.Insert(int32(meshType))
	}

	if representedMeshTypes.Len() > 1 {
		return OnlyHomogenousVirtualMeshesSupported(vm.Name, vm.Namespace, representedMeshTypes.Len())
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
