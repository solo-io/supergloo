package decider

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/decider/strategies"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	"go.uber.org/zap"
)

var (
	UnsupportedFederationMode = "Unsupported federation_mode"
	ErrorUpdatingMeshServices = func(err error) string {
		return fmt.Sprintf("Error while updating mesh services' federation metadata: %s", err.Error())
	}
	ErrorLoadingMeshMetadata = func(err error) string {
		return fmt.Sprintf("Error while loading mesh metadata to determine federation resolution: %s", err.Error())
	}
)

func NewFederationSnapshotListener(decider FederationDecider) FederationDeciderSnapshotListener {
	return &snapshot.MeshNetworkingSnapshotListenerFunc{
		OnSync: func(ctx context.Context, snap *snapshot.MeshNetworkingSnapshot) {
			decider.DecideFederation(ctx, snap)
		},
	}
}

type FederationDeciderSnapshotListener snapshot.MeshNetworkingSnapshotListener

type FederationDecider interface {
	DecideFederation(ctx context.Context, snap *snapshot.MeshNetworkingSnapshot)
}

func NewFederationDecider(
	meshServiceClient discovery_core.MeshServiceClient,
	meshClient discovery_core.MeshClient,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
	federationStrategyChooser strategies.FederationStrategyChooser,
) FederationDecider {
	return &federationDecider{
		meshServiceClient:         meshServiceClient,
		meshClient:                meshClient,
		virtualMeshClient:         virtualMeshClient,
		federationStrategyChooser: federationStrategyChooser,
	}
}

type federationDecider struct {
	meshServiceClient         discovery_core.MeshServiceClient
	meshClient                discovery_core.MeshClient
	virtualMeshClient         zephyr_networking.VirtualMeshClient
	federationStrategyChooser strategies.FederationStrategyChooser
}

func (f *federationDecider) DecideFederation(ctx context.Context, networkingSnapshot *snapshot.MeshNetworkingSnapshot) {
	logger := contextutils.LoggerFrom(ctx)

	perMeshMetadata, errorReports := strategies.BuildPerMeshMetadataFromSnapshot(ctx, networkingSnapshot, f.meshClient)

	// log and update the status just for the ones that failed, then continue
	if len(errorReports) > 0 {
		for _, failedVirtualMeshReport := range errorReports {
			failedVirtualMeshReport.VirtualMesh.Status.FederationStatus = &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_PROCESSING_ERROR,
				Message: ErrorLoadingMeshMetadata(failedVirtualMeshReport.Err),
			}

			f.updateVirtualMeshStatus(ctx, failedVirtualMeshReport.VirtualMesh)

			logger.Errorf("Failed to load federation data for virtual mesh %s: %s", failedVirtualMeshReport.VirtualMesh.GetName(), failedVirtualMeshReport.Err.Error())
		}
	}

	for _, vm := range perMeshMetadata.ResolvedVirtualMeshs {
		f.federateVirtualMesh(
			ctx,
			vm,
			perMeshMetadata,
		)
	}
}

func (f *federationDecider) federateVirtualMesh(
	ctx context.Context,
	vm *networking_v1alpha1.VirtualMesh,
	perMeshMetadata strategies.PerMeshMetadata,
) {
	logger := contextutils.LoggerFrom(ctx)

	// if federation has not been explicitly set by the user, this expression will default the federation mode
	// to PERMISSIVE, which probably isn't what we want long-term. Tracking that future change here:
	// https://github.com/solo-io/mesh-projects/issues/222
	federationMode := vm.Spec.GetFederation().GetMode()

	// determine what strategy we should use to federate
	federationStrategy, err := f.federationStrategyChooser(federationMode, f.meshServiceClient)
	if err != nil {
		vm.Status.FederationStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: UnsupportedFederationMode,
		}
		f.updateVirtualMeshStatus(ctx, vm)
		return
	}

	// actually write our federation decision to the mesh services
	err = federationStrategy.WriteFederationToServices(ctx, vm, perMeshMetadata.MeshNameToMetadata)
	if err == nil {
		vm.Status.FederationStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
	} else {
		logger.Debugf("Recording error to virtual mesh %s.%s", vm.Name, vm.Namespace, zap.Error(err))
		vm.Status.FederationStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_PROCESSING_ERROR,
			Message: ErrorUpdatingMeshServices(err),
		}
	}

	f.updateVirtualMeshStatus(ctx, vm)
}

// once the virtual mesh has had its federation status updated, call this function to write it into the cluster
func (f *federationDecider) updateVirtualMeshStatus(ctx context.Context, virtualMesh *networking_v1alpha1.VirtualMesh) {
	logger := contextutils.LoggerFrom(ctx)

	err := f.virtualMeshClient.UpdateStatus(ctx, virtualMesh)
	if err != nil {
		logger.Errorf("Error updating federation status on virtual mesh %s.%s", virtualMesh.Name, virtualMesh.Namespace)
	}
}
