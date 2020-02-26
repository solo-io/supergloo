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

func NewFederationDecider(
	meshServiceClient discovery_core.MeshServiceClient,
	meshClient discovery_core.MeshClient,
	meshGroupClient zephyr_networking.MeshGroupClient,
	federationStrategyChooser strategies.FederationStrategyChooser,
) FederationDecider {
	return &federationDecider{
		meshServiceClient:         meshServiceClient,
		meshClient:                meshClient,
		meshGroupClient:           meshGroupClient,
		federationStrategyChooser: federationStrategyChooser,
	}
}

type federationDecider struct {
	meshServiceClient         discovery_core.MeshServiceClient
	meshClient                discovery_core.MeshClient
	meshGroupClient           zephyr_networking.MeshGroupClient
	federationStrategyChooser strategies.FederationStrategyChooser
}

func (f *federationDecider) DecideFederation(ctx context.Context, networkingSnapshot snapshot.MeshNetworkingSnapshot) {
	logger := contextutils.LoggerFrom(ctx)

	perMeshMetadata, errorReports := strategies.BuildPerMeshMetadataFromSnapshot(ctx, networkingSnapshot, f.meshClient)

	// log and update the status just for the ones that failed, then continue
	if len(errorReports) > 0 {
		for _, failedGroupReport := range errorReports {
			failedGroupReport.MeshGroup.Status.FederationStatus = &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_PROCESSING_ERROR,
				Message: ErrorLoadingMeshMetadata(failedGroupReport.Err),
			}

			f.updateMeshGroupStatus(ctx, failedGroupReport.MeshGroup)

			logger.Errorf("Failed to load federation data for group %s: %s", failedGroupReport.MeshGroup.GetName(), failedGroupReport.Err.Error())
		}
	}

	for _, group := range perMeshMetadata.ResolvedMeshGroups {
		f.federateGroup(
			ctx,
			group,
			perMeshMetadata,
		)
	}
}

func (f *federationDecider) federateGroup(
	ctx context.Context,
	group *networking_v1alpha1.MeshGroup,
	perMeshMetadata strategies.PerMeshMetadata,
) {
	logger := contextutils.LoggerFrom(ctx)

	// if federation has not been explicitly set by the user, this expression will default the federation mode
	// to PERMISSIVE, which probably isn't what we want long-term. Tracking that future change here:
	// https://github.com/solo-io/mesh-projects/issues/222
	federationMode := group.Spec.GetFederation().GetMode()

	// determine what strategy we should use to federate
	federationStrategy, err := f.federationStrategyChooser(federationMode, f.meshServiceClient)
	if err != nil {
		group.Status.FederationStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: UnsupportedFederationMode,
		}
		f.updateMeshGroupStatus(ctx, group)
		return
	}

	// actually write our federation decision to the mesh services
	err = federationStrategy.WriteFederationToServices(ctx, group, perMeshMetadata.MeshNameToMetadata)
	if err == nil {
		group.Status.FederationStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
	} else {
		logger.Error("Recording error to mesh group %+v: %+v", group.ObjectMeta, err)
		group.Status.FederationStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_PROCESSING_ERROR,
			Message: ErrorUpdatingMeshServices(err),
		}
	}

	f.updateMeshGroupStatus(ctx, group)
}

// once the mesh group has had its federation status updated, call this function to write it into the cluster
func (f *federationDecider) updateMeshGroupStatus(ctx context.Context, meshGroup *networking_v1alpha1.MeshGroup) {
	logger := contextutils.LoggerFrom(ctx)

	err := f.meshGroupClient.UpdateStatus(ctx, meshGroup)
	if err != nil {
		logger.Errorf("Error updating federation status on mesh group %+v", meshGroup.ObjectMeta)
	}
}
