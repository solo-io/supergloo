package internal

import (
	"context"
	"fmt"

	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
)

func NewFederationDecisionCheck() healthcheck_types.HealthCheck {
	return &federationDecisionCheck{}
}

type federationDecisionCheck struct{}

func (*federationDecisionCheck) GetDescription() string {
	return "federation decisions have been written to MeshServices"
}

func (s *federationDecisionCheck) Run(ctx context.Context, installNamespace string, clients healthcheck_types.Clients) (failure *healthcheck_types.RunFailure, checkApplies bool) {
	meshServices, err := clients.MeshServiceClient.ListMeshService(ctx)
	if err != nil {
		return &healthcheck_types.RunFailure{
			ErrorMessage: GenericCheckFailed(err).Error(),
		}, true
	}

	if len(meshServices.Items) == 0 {
		return nil, false
	}

	// don't bother reporting a status here if no services have been federated anywhere
	federatedServiceExists := false
	for _, meshService := range meshServices.Items {
		federationStatus := meshService.Status.FederationStatus
		if federationStatus == nil {
			// we have not performed federation on this service yet
			continue
		}

		federatedServiceExists = true
		if federationStatus.GetState() != smh_core_types.Status_ACCEPTED {
			return &healthcheck_types.RunFailure{
				ErrorMessage: FederationRecordingHasFailed(meshService.GetName(), meshService.GetNamespace(), federationStatus.State).Error(),
				Hint:         fmt.Sprintf("get details from the failing MeshService: `kubectl -n %s get meshservice %s -oyaml`", installNamespace, meshService.GetName()),
			}, true
		}
	}

	// the check only applies if we have federated services
	return nil, federatedServiceExists
}
