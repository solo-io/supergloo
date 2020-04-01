package access_policy_enforcer

import (
	"context"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/clients"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/pkg/logging"
)

type enforcerLoop struct {
	virtualMeshController controller.VirtualMeshController
	virtualMeshClient     zephyr_networking.VirtualMeshClient
	meshClient            zephyr_discovery.MeshClient
	meshEnforcers         []AccessPolicyMeshEnforcer
}

func NewEnforcerLoop(
	virtualMeshController controller.VirtualMeshController,
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
	return e.virtualMeshController.AddEventHandler(ctx, &controller.VirtualMeshEventHandlerFuncs{
		OnCreate: func(virtualMesh *networking_v1alpha1.VirtualMesh) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, virtualMesh)
			logger.Debugf("Handling event: %+v", virtualMesh)
			err := e.enforceGlobalAccessControl(ctx, virtualMesh)
			err = e.setStatus(ctx, virtualMesh, err)
			if err != nil {
				logger.Errorw("Error while handling VirtualMesh create event", err)
			}
			return nil
		},
		OnUpdate: func(_, virtualMesh *networking_v1alpha1.VirtualMesh) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, virtualMesh)
			logger.Debugf("Handling event: %+v", virtualMesh)
			err := e.enforceGlobalAccessControl(ctx, virtualMesh)
			err = e.setStatus(ctx, virtualMesh, err)
			if err != nil {
				logger.Errorw("Error while handling VirtualMesh create event", err)
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
			if err := meshEnforcer.StartEnforcing(ctx, meshes); err != nil {
				return err
			}
		} else {
			if err := meshEnforcer.StopEnforcing(ctx, meshes); err != nil {
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
		mesh, err := e.meshClient.Get(ctx, clients.ResourceRefToObjectKey(meshRef))
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
		virtualMesh.Status.AccessControlEnforcementStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_PROCESSING_ERROR,
			Message: err.Error(),
		}
	} else {
		virtualMesh.Status.AccessControlEnforcementStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
	}
	return e.virtualMeshClient.UpdateStatus(ctx, virtualMesh)
}
