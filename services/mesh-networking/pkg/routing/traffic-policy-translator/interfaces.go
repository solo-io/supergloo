package traffic_policy_translator

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type TrafficPolicyMeshTranslator interface {
	// The name which will be used to identify the translator in the logs
	Name() string
	// translate a TrafficPolicy into mesh-specific resources and upsert
	// the presence of a TranslatorError indicates an error during translation
	TranslateTrafficPolicy(
		ctx context.Context,
		meshService *zephyr_discovery.MeshService,
		mesh *zephyr_discovery.Mesh,
		mergedTrafficPolicy []*zephyr_networking.TrafficPolicy,
	) *types.TrafficPolicyStatus_TranslatorError
}

type TrafficPolicyTranslatorLoop interface {
	Start(ctx context.Context) error
}
