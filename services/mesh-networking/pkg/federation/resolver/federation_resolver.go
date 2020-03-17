package resolver

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controllers "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/dns"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/resolver/meshes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	FailedToFederateService = func(err error, meshService *discovery_v1alpha1.MeshService, meshWorkloadRef *core_types.ResourceRef) string {
		return fmt.Sprintf("Could not federate service %+v to mesh workload %+v: %s", meshService.ObjectMeta, meshWorkloadRef, err.Error())
	}
)

type PerMeshFederationClients struct {
	Istio meshes.MeshFederationClient
}

func NewFederationResolver(
	meshClient zephyr_discovery.MeshClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
	perMeshFederationClients PerMeshFederationClients,
) FederationResolver {
	return &federationResolver{
		meshClient:               meshClient,
		meshWorkloadClient:       meshWorkloadClient,
		meshServiceClient:        meshServiceClient,
		virtualMeshClient:        virtualMeshClient,
		perMeshFederationClients: perMeshFederationClients,
	}
}

type federationResolver struct {
	meshClient               zephyr_discovery.MeshClient
	meshWorkloadClient       zephyr_discovery.MeshWorkloadClient
	meshServiceClient        zephyr_discovery.MeshServiceClient
	virtualMeshClient        zephyr_networking.VirtualMeshClient
	perMeshFederationClients PerMeshFederationClients
}

func (f *federationResolver) Start(ctx context.Context, meshServiceController discovery_controllers.MeshServiceController) {
	meshServiceController.AddEventHandler(ctx, &discovery_controllers.MeshServiceEventHandlerFuncs{
		OnCreate: func(obj *discovery_v1alpha1.MeshService) error {
			return f.handleServiceUpsert(ctx, obj)
		},
		OnUpdate: func(old, new *discovery_v1alpha1.MeshService) error {
			// for status-only updates, do nothing
			// this is important to ensure that we eventually get into a consistent state, as
			// this component is also responsible for writing mesh service statuses
			if old.Spec.Equal(new.Spec) {
				return nil
			}

			return f.handleServiceUpsert(ctx, new)
		},
		OnDelete: func(_ *discovery_v1alpha1.MeshService) error {
			// ignoring delete
			// https://github.com/solo-io/mesh-projects/issues/169
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

	for _, federatedToWorkloadRef := range federationConfig.FederatedToWorkloads {
		err := f.federateToRemoteWorkload(ctx, meshService, federatedToWorkloadRef)

		var federationStatus *core_types.ComputedStatus
		if err != nil {
			message := FailedToFederateService(err, meshService, federatedToWorkloadRef)
			logger.Errorf(message)

			federationStatus = &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_PROCESSING_ERROR,
				Message: message,
			}
		} else {
			federationStatus = &core_types.ComputedStatus{
				Status: core_types.ComputedStatus_ACCEPTED,
			}
		}

		meshService.Status.FederationStatus = federationStatus
		err = f.meshServiceClient.UpdateStatus(ctx, meshService)
		if err != nil {
			logger.Errorf("Failed to update service status: %+v", err)
		}
	}

	return nil
}

func (f *federationResolver) federateToRemoteWorkload(ctx context.Context, meshService *discovery_v1alpha1.MeshService, meshWorkloadRef *core_types.ResourceRef) error {
	workload, err := f.meshWorkloadClient.Get(ctx, client.ObjectKey{
		Name:      meshWorkloadRef.GetName(),
		Namespace: meshWorkloadRef.GetNamespace(),
	})
	if err != nil {
		return eris.Wrapf(err, "Could not load federated MeshWorkload metadata for service %+v", meshService.ObjectMeta)
	}

	meshForWorkload, err := f.meshClient.Get(ctx, client.ObjectKey{
		Name:      workload.Spec.GetMesh().GetName(),
		Namespace: workload.Spec.GetMesh().GetNamespace(),
	})
	if err != nil {
		return eris.Wrapf(err, "Could not load mesh for MeshWorkload %+v", workload.ObjectMeta)
	}

	meshForService, err := f.meshClient.Get(ctx, client.ObjectKey{
		Name:      meshService.Spec.GetMesh().GetName(),
		Namespace: meshService.Spec.GetMesh().GetNamespace(),
	})
	if err != nil {
		return eris.Wrapf(err, "Could not load mesh for MeshService %+v", meshService.ObjectMeta)
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
		eap, err = f.perMeshFederationClients.Istio.FederateServiceSide(ctx, virtualMesh, meshService)
	default:
		err = eris.Errorf("Unsupported mesh type for federation: %+v", meshForWorkload.Spec.MeshType)
	}

	if err != nil {
		return err
	}

	// set up gateway resources on the client cluster
	switch meshForWorkload.Spec.GetMeshType().(type) {
	case *types.MeshSpec_Istio:
		return f.perMeshFederationClients.Istio.FederateClientSide(
			ctx,
			eap,
			meshService,
			workload,
		)
	default:
		return eris.Errorf("Unsupported mesh type for federation: %+v", meshForWorkload.Spec.MeshType)
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
