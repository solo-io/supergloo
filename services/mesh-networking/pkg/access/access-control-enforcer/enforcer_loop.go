package access_policy_enforcer

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"go.uber.org/zap"
)

type enforcerLoop struct {
	virtualMeshEventWatcher zephyr_networking_controller.VirtualMeshEventWatcher
	virtualMeshClient       zephyr_networking.VirtualMeshClient
	meshClient              zephyr_discovery.MeshClient
	meshEnforcers           []AccessPolicyMeshEnforcer
}

func NewEnforcerLoop(
	virtualMeshEventWatcher zephyr_networking_controller.VirtualMeshEventWatcher,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
	meshClient zephyr_discovery.MeshClient,
	meshEnforcers []AccessPolicyMeshEnforcer,
) AccessPolicyEnforcerLoop {
	return &enforcerLoop{
		virtualMeshEventWatcher: virtualMeshEventWatcher,
		virtualMeshClient:       virtualMeshClient,
		meshClient:              meshClient,
		meshEnforcers:           meshEnforcers,
	}
}

func (e *enforcerLoop) Start(ctx context.Context) error {
	return e.virtualMeshEventWatcher.AddEventHandler(ctx, &zephyr_networking_controller.VirtualMeshEventHandlerFuncs{
		OnCreate: func(obj *zephyr_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.CreateEvent, obj)
			logger.Debugw("event handler enter",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			err := e.enforceGlobalAccessControl(ctx, obj)
			err = e.setStatus(ctx, obj, err)
			if err != nil {
				logger.Errorw("Error while handling VirtualMesh create event", err)
			}
			return nil
		},
		OnUpdate: func(old, new *zephyr_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.UpdateEvent, new)
			logger.Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			err := e.enforceGlobalAccessControl(ctx, new)
			err = e.setStatus(ctx, new, err)
			if err != nil {
				logger.Errorw("Error while handling VirtualMesh update event", err)
			}
			return nil
		},
		OnDelete: func(virtualMesh *zephyr_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.DeleteEvent, virtualMesh)
			logger.Debugw("event handler enter",
				zap.Any("spec", virtualMesh.Spec),
				zap.Any("status", virtualMesh.Status),
			)

			// manually set this to false so that things get cleaned up
			virtualMesh.Spec.EnforceAccessControl = false

			// TODO https://github.com/solo-io/service-mesh-hub/issues/650 - we probably want to introduce some defensive retries into our code
			err := e.enforceGlobalAccessControl(ctx, virtualMesh)
			if err != nil {
				logger.Errorf("%+v", err)
			}

			return nil
		},
		OnGeneric: func(virtualMesh *zephyr_networking.VirtualMesh) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.GenericEvent, virtualMesh)
			logger.Debugf("Ignoring event: %+v", virtualMesh)
			return nil
		},
	})
}

func (e *enforcerLoop) enforceGlobalAccessControl(
	ctx context.Context,
	virtualMesh *zephyr_networking.VirtualMesh,
) error {
	meshes, err := e.fetchMeshes(ctx, virtualMesh)
	if err != nil {
		return err
	}
	for _, meshEnforcer := range e.meshEnforcers {
		if virtualMesh.Spec.GetEnforceAccessControl() {
			if err = meshEnforcer.StartEnforcing(
				contextutils.WithLogger(ctx, meshEnforcer.Name()),
				meshes,
			); err != nil {
				return err
			}
		} else {
			if err = meshEnforcer.StopEnforcing(
				contextutils.WithLogger(ctx, meshEnforcer.Name()),
				meshes,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *enforcerLoop) fetchMeshes(
	ctx context.Context,
	virtualMesh *zephyr_networking.VirtualMesh,
) ([]*zephyr_discovery.Mesh, error) {
	var meshes []*zephyr_discovery.Mesh
	for _, meshRef := range virtualMesh.Spec.GetMeshes() {
		mesh, err := e.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshRef))
		if err != nil {
			return nil, err
		}
		meshes = append(meshes, mesh)
	}
	return meshes, nil
}

func (e *enforcerLoop) setStatus(
	ctx context.Context,
	virtualMesh *zephyr_networking.VirtualMesh,
	err error,
) error {
	if err != nil {
		virtualMesh.Status.AccessControlEnforcementStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
	} else {
		virtualMesh.Status.AccessControlEnforcementStatus = &zephyr_core_types.Status{
			State: zephyr_core_types.Status_ACCEPTED,
		}
	}
	return e.virtualMeshClient.UpdateVirtualMeshStatus(ctx, virtualMesh)
}
