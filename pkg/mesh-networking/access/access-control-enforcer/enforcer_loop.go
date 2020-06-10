package access_policy_enforcer

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/pkg/common/access-control/enforcer"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type enforcerLoop struct {
	virtualMeshEventWatcher smh_networking_controller.VirtualMeshEventWatcher
	virtualMeshClient       smh_networking.VirtualMeshClient
	meshClient              smh_discovery.MeshClient
	meshEnforcers           []access_control_enforcer.AccessPolicyMeshEnforcer
}

func NewEnforcerLoop(
	virtualMeshEventWatcher smh_networking_controller.VirtualMeshEventWatcher,
	virtualMeshClient smh_networking.VirtualMeshClient,
	meshClient smh_discovery.MeshClient,
	meshEnforcers []access_control_enforcer.AccessPolicyMeshEnforcer,
) AccessPolicyEnforcerLoop {
	return &enforcerLoop{
		virtualMeshEventWatcher: virtualMeshEventWatcher,
		virtualMeshClient:       virtualMeshClient,
		meshClient:              meshClient,
		meshEnforcers:           meshEnforcers,
	}
}

func (e *enforcerLoop) Start(ctx context.Context) error {
	return e.virtualMeshEventWatcher.AddEventHandler(ctx, &smh_networking_controller.VirtualMeshEventHandlerFuncs{
		OnCreate: func(obj *smh_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.CreateEvent, obj)
			logger.Debugw("event handler enter",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			return e.reconcile(ctx, container_runtime.CreateEvent)
		},
		OnUpdate: func(old, new *smh_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.UpdateEvent, new)
			logger.Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			return e.reconcile(ctx, container_runtime.UpdateEvent)
		},
		OnDelete: func(obj *smh_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.DeleteEvent, obj)
			logger.Debugw("event handler enter",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			return e.reconcile(ctx, container_runtime.DeleteEvent)
		},
		OnGeneric: func(virtualMesh *smh_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.GenericEvent, virtualMesh)
			logger.Debugf("Ignoring event: %+v", virtualMesh)
			return nil
		},
	})
}

func (e *enforcerLoop) reconcile(ctx context.Context, eventType container_runtime.EventType) error {
	vmList, err := e.virtualMeshClient.ListVirtualMesh(ctx)
	if err != nil {
		return err
	}
	for _, vm := range vmList.Items {
		vm := vm
		logger := container_runtime.BuildEventLogger(ctx, eventType, &vm)
		err = e.enforceGlobalAccessControl(ctx, &vm)
		if err != nil {
			logger.Errorf("Error while handling VirtualMesh event: %+v", err)
		}
		if eventType != container_runtime.DeleteEvent {
			statusErr := e.setStatus(ctx, &vm, err)
			if statusErr != nil {
				logger.Errorf("Error while settings status for VirtualMesh: %+v", statusErr)
				return statusErr
			}
		}
	}
	// TODO: temporary logic, see comment on function definition
	e.cleanupAppmeshResources(ctx)
	return err
}

// Temporary logic: currently if an Appmesh instance is grouped into a VirtualMesh,
// translation occurs implicitly to allow communication between all workloads and services.
// We have an outstanding issue, https://github.com/solo-io/service-mesh-hub/issues/750, for extending
// the SMH API to support configuration of sidecar traffic mediation.
// So for now, we need a way to delete implicitly translated Appmesh resources if a Mesh is removed from a VirtualMesh.
func (e *enforcerLoop) cleanupAppmeshResources(ctx context.Context) error {
	meshList, err := e.meshClient.ListMesh(ctx)
	if err != nil {
		return err
	}
	allMeshes := smh_discovery_sets.NewMeshSet()
	meshesInVirtualMesh := smh_discovery_sets.NewMeshSet()
	for _, mesh := range meshList.Items {
		mesh := mesh
		if mesh.Spec.GetAwsAppMesh() == nil {
			continue
		}
		allMeshes.Insert(&mesh)
	}
	virtualMeshList, err := e.virtualMeshClient.ListVirtualMesh(ctx)
	if err != nil {
		return err
	}
	for _, vm := range virtualMeshList.Items {
		vm := vm
		for _, meshRef := range vm.Spec.GetMeshes() {
			meshesInVirtualMesh.Insert(&smh_discovery.Mesh{
				ObjectMeta: v1.ObjectMeta{
					Name:      meshRef.GetName(),
					Namespace: meshRef.GetNamespace(),
				},
			})
		}
	}
	meshesWithoutVM := allMeshes.Difference(meshesInVirtualMesh).List()
	for _, mesh := range meshesWithoutVM {
		for _, meshEnforcer := range e.meshEnforcers {
			err = meshEnforcer.ReconcileAccessControl(
				contextutils.WithLogger(ctx, meshEnforcer.Name()),
				mesh,
				nil,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// If enforceMeshDefault, ignore user declared enforce_access_control setting and use mesh-specific default as defined on the
// VirtualMesh API, used for VM deletion to clean up mesh resources.
// If virtualMesh is nil, enforce the mesh specific default.
func (e *enforcerLoop) enforceGlobalAccessControl(
	ctx context.Context,
	virtualMesh *smh_networking.VirtualMesh,
) error {
	meshes, err := e.fetchMeshes(ctx, virtualMesh)
	if err != nil {
		return err
	}
	if len(meshes) == 0 {
		return nil
	}
	for _, mesh := range meshes {
		for _, meshEnforcer := range e.meshEnforcers {
			err = meshEnforcer.ReconcileAccessControl(
				contextutils.WithLogger(ctx, meshEnforcer.Name()),
				mesh,
				virtualMesh,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *enforcerLoop) fetchMeshes(
	ctx context.Context,
	virtualMesh *smh_networking.VirtualMesh,
) ([]*smh_discovery.Mesh, error) {
	var meshes []*smh_discovery.Mesh
	for _, meshRef := range virtualMesh.Spec.GetMeshes() {
		mesh, err := e.meshClient.GetMesh(ctx, selection.ResourceRefToObjectKey(meshRef))
		if err != nil {
			return nil, err
		}
		meshes = append(meshes, mesh)
	}
	return meshes, nil
}

func (e *enforcerLoop) setStatus(
	ctx context.Context,
	virtualMesh *smh_networking.VirtualMesh,
	err error,
) error {
	if err != nil {
		virtualMesh.Status.AccessControlEnforcementStatus = &smh_core_types.Status{
			State:   smh_core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
	} else {
		virtualMesh.Status.AccessControlEnforcementStatus = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
	}
	return e.virtualMeshClient.UpdateVirtualMeshStatus(ctx, virtualMesh)
}
