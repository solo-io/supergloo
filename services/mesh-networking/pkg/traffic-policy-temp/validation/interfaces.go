package traffic_policy_validation

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

type Validator interface {
	// always returns a non-nil Status that should be written to the cluster
	// if validation failed, that status will also be returned with the concrete error that occurred so it can be logged
	ValidateTrafficPolicy(trafficPolicy *zephyr_networking.TrafficPolicy, allMeshServices []*zephyr_discovery.MeshService) (*zephyr_core_types.Status, error)
}

type ValidationLoop interface {
	RunOnce(ctx context.Context) error
}
