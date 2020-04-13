package access_policy_enforcer

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"go.uber.org/zap"
)

type enforcerLoop struct {
	virtualMeshController networking_controller.VirtualMeshEventWatcher
	virtualMeshClient     zephyr_networking.VirtualMeshClient
	meshClient            zephyr_discovery.MeshClient
	meshEnforcers         []AccessPolicyMeshEnforcer
}

func NewEnforcerLoop(
	virtualMeshController networking_controller.VirtualMeshEventWatcher,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
	meshClient zephyr_discovery.MeshClient,
	meshEnforcers []AccessPolicyMeshEnforcer,
) AccessPolicyEnforcerLoop {
	return &enforcerLoop{
		virtualMeshController: virtualMeshController,
		virtualMeshClient:     virtualMeshClient,
		meshClient:            meshClient,
		meshEnforcers:         meshEnforcers,
	}
}

func (e *enforcerLoop) Start(ctx context.Context) error {
	return e.virtualMeshController.AddEventHandler(ctx, &networking_controller.VirtualMeshEventHandlerFuncs{
		OnCreate: func(obj *networking_v1alpha1.VirtualMesh) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, obj)
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
		OnUpdate: func(old, new *networking_v1alpha1.VirtualMesh) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
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
		OnDelete: func(virtualMesh *networking_v1alpha1.VirtualMesh) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, virtualMesh)
			logger.Debugf("Ignoring event: %+v", virtualMesh)
			return nil
		},
		OnGeneric: func(virtualMesh *networking_v1alpha1.VirtualMesh) error {
			logger := logging.BuildEventLogger(ctx, logging.GenericEvent, virtualMesh)
			logger.Debugf("Ignoring event: %+v", virtualMesh)
			return nil
		},
	})
}

func (e *enforcerLoop) enforceGlobalAccessControl(
	ctx context.Context,
	virtualMesh *networking_v1alpha1.VirtualMesh,
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
	virtualMesh *networking_v1alpha1.VirtualMesh,
) ([]*discovery_v1alpha1.Mesh, error) {
	var meshes []*discovery_v1alpha1.Mesh
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
	virtualMesh *networking_v1alpha1.VirtualMesh,
	err error,
) error {
	if err != nil {
		virtualMesh.Status.AccessControlEnforcementStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
	} else {
		virtualMesh.Status.AccessControlEnforcementStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
	}
	return e.virtualMeshClient.UpdateVirtualMeshStatus(ctx, virtualMesh)
}
