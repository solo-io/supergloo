package traffic_policy_validation

import (
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// do data validation on a Traffic Policy; e.g., ensure that its retry attempts are non-negative, that it references real services, etc.
type Validator interface {
	// always returns a non-nil Status that should be written to the cluster
	// if validation failed, the concrete validation error that occurred will be returned with that non-nil status so it can be logged
	ValidateTrafficPolicy(trafficPolicy *smh_networking.TrafficPolicy, allMeshServices []*smh_discovery.MeshService) (*smh_core_types.Status, error)
}
