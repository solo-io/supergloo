package resolver

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/dns"
	istio_federation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/resolver/meshes/istio"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FailedToFederateServiceMessage = "failed to federate mesh service to mesh workload"
)

var (
	FailedToFederateServices = func(
		meshService *discovery_v1alpha1.MeshService,
		meshWorkloadRefs []*core_types.ResourceRef) string {
		return fmt.Sprintf("Could not federate service %s.%s to mesh workloads %+v. Check logs for details",
			meshService.Name,
			meshService.Namespace,
			meshWorkloadRefs,
		)
	}
)

type PerMeshFederationClients struct {
	Istio istio_federation.IstioFederationClient
}

func NewPerMeshFederationClients(istio istio_federation.IstioFederationClient) PerMeshFederationClients {
	return PerMeshFederationClients{Istio: istio}
}

func NewFederationResolver(
	meshClient zephyr_discovery.MeshClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
	perMeshFederationClients PerMeshFederationClients,
	meshServiceController discovery_controllers.MeshServiceController,
) FederationResolver {
	return &federationResolver{
		meshClient:               meshClient,
		meshWorkloadClient:       meshWorkloadClient,
		meshServiceClient:        meshServiceClient,
		virtualMeshClient:        virtualMeshClient,
		perMeshFederationClients: perMeshFederationClients,
		meshServiceController:    meshServiceController,
	}
}

type federationResolver struct {
	meshServiceController    discovery_controllers.MeshServiceController
	meshClient               zephyr_discovery.MeshClient
	meshWorkloadClient       zephyr_discovery.MeshWorkloadClient
	meshServiceClient        zephyr_discovery.MeshServiceClient
	virtualMeshClient        zephyr_networking.VirtualMeshClient
	perMeshFederationClients PerMeshFederationClients
}

func (f *federationResolver) Start(ctx context.Context) error {
	return f.meshServiceController.AddEventHandler(ctx, &discovery_controllers.MeshServiceEventHandlerFuncs{
		OnCreate: func(obj *discovery_v1alpha1.MeshService) error {
			eventCtx := logging.EventContext(ctx, logging.CreateEvent, obj)
			contextutils.LoggerFrom(eventCtx).Debugw("event handler enter",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			return f.handleServiceUpsert(eventCtx, obj)
		},
		OnUpdate: func(old, new *discovery_v1alpha1.MeshService) error {
			eventCtx := logging.EventContext(ctx, logging.CreateEvent, new)
			// for status-only updates, do nothing
			// this is important to ensure that we eventually get into a consistent state, as
			// this component is also responsible for writing mesh service statuses
			contextutils.LoggerFrom(eventCtx).Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			if old.Spec.Equal(new.Spec) {
				return nil
			}

			return f.handleServiceUpsert(eventCtx, new)
		},
		OnDelete: func(_ *discovery_v1alpha1.MeshService) error {
			// ignoring delete
			// https://github.com/solo-io/service-mesh-hub/issues/169
			return nil
		},
		OnGeneric: nil,
	})
}

// handle services that get created or updated
func (f *federationResolver) handleServiceUpsert(ctx context.Context, meshService *discovery_v1alpha1.MeshService) error {
	logger := contextutils.LoggerFrom(ctx)

	federationConfig := meshService.Spec.GetFederation()
	if federationConfig == nil {
		return nil
	}
	var failedFederations []*core_types.ResourceRef
	for _, federatedToWorkloadRef := range federationConfig.FederatedToWorkloads {
		if err := f.federateToRemoteWorkload(ctx, meshService, federatedToWorkloadRef); err != nil {
			failedFederations = append(failedFederations, federatedToWorkloadRef)
			logger.Warnw(FailedToFederateServiceMessage,
				zap.String("mesh_workload", fmt.Sprintf("%s.%s", federatedToWorkloadRef.GetName(), federatedToWorkloadRef.GetNamespace())),
				zap.String("mesh_service", fmt.Sprintf("%s.%s", meshService.GetName(), meshService.GetNamespace())),
				zap.Error(err))
		}
	}

	var federationStatus *core_types.Status
	if len(failedFederations) > 0 {
		federationStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: FailedToFederateServices(meshService, failedFederations),
		}
	} else {
		federationStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
	}

	// If the status is the same as the current, do not attempt to update
	if meshService.Status.FederationStatus.Equal(federationStatus) {
		logger.Debugw("federation status is equal, not updating",
			zap.Any("existing_status", meshService.Status.FederationStatus),
			zap.Any("new_status", federationStatus),
		)
		return nil
	}
	logger.Debugw("federation status equal,updating",
		zap.Any("existing_status", meshService.Status.FederationStatus),
		zap.Any("new_status", federationStatus),
	)

	meshService.Status.FederationStatus = federationStatus
	err := f.meshServiceClient.UpdateStatus(ctx, meshService)
	if err != nil {
		logger.Warnw("Failed to update service status", zap.Error(err))
	}

	return nil
}

func (f *federationResolver) federateToRemoteWorkload(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	meshWorkloadRef *core_types.ResourceRef,
) error {
	workload, err := f.meshWorkloadClient.Get(ctx, client.ObjectKey{
		Name:      meshWorkloadRef.GetName(),
		Namespace: meshWorkloadRef.GetNamespace(),
	})
	if err != nil {
		return eris.Wrapf(err, "Could not load federated MeshWorkload metadata for service %s.%s",
			meshService.GetNamespace(), meshService.GetNamespace())
	}

	meshForWorkload, err := f.meshClient.Get(ctx, client.ObjectKey{
		Name:      workload.Spec.GetMesh().GetName(),
		Namespace: workload.Spec.GetMesh().GetNamespace(),
	})
	if err != nil {
		return eris.Wrapf(err, "Could not load mesh for MeshWorkload %s.%s",
			meshService.GetNamespace(),
			meshService.GetNamespace(),
		)
	}

	meshForService, err := f.meshClient.Get(ctx, client.ObjectKey{
		Name:      meshService.Spec.GetMesh().GetName(),
		Namespace: meshService.Spec.GetMesh().GetNamespace(),
	})
	if err != nil {
		return eris.Wrapf(err, "Could not load mesh for MeshService %s.%s",
			meshService.ObjectMeta.Name,
			meshService.ObjectMeta.Namespace,
		)
	}

	virtualMesh, err := f.getVirtualMeshContainingService(ctx, meshForService)
	if err != nil {
		return err
	}

	var (
		eap dns.ExternalAccessPoint
	)
	// set up gateway resources on the target cluster
	switch meshForService.Spec.GetMeshType().(type) {
	case *types.MeshSpec_Istio:
		eap, err = f.perMeshFederationClients.Istio.FederateServiceSide(
			contextutils.WithLogger(ctx, "istio"),
			virtualMesh,
			meshService,
		)
	default:
		err = eris.Errorf("Unsupported mesh type for federation: %T", meshForWorkload.Spec.MeshType)
	}

	if err != nil {
		return err
	}

	// set up gateway resources on the client cluster
	switch meshForWorkload.Spec.GetMeshType().(type) {
	case *types.MeshSpec_Istio:
		return f.perMeshFederationClients.Istio.FederateClientSide(
			contextutils.WithLogger(ctx, "istio"),
			eap,
			meshService,
			workload,
		)
	default:
		return eris.Errorf("Unsupported mesh type for federation: %T", meshForWorkload.Spec.MeshType)
	}
}

func (f *federationResolver) getVirtualMeshContainingService(
	ctx context.Context,
	meshForService *discovery_v1alpha1.Mesh,
) (*networking_v1alpha1.VirtualMesh, error) {
	virtualMeshs, err := f.virtualMeshClient.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, virtualMesh := range virtualMeshs.Items {
		for _, mesh := range virtualMesh.Spec.GetMeshes() {
			if mesh.GetName() == meshForService.GetName() && mesh.GetNamespace() == meshForService.GetNamespace() {
				return &virtualMesh, nil
			}
		}
	}

	return nil, eris.Errorf("No virtual mesh found containing mesh %s", meshForService.GetName())
}
