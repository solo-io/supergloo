package traffic_policy_translator

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type TrafficPolicyMeshTranslator interface {
	// translate a TrafficPolicy into mesh-specific resources and upsert
	// the presence of a TranslatorError indicates an error during translation
	TranslateTrafficPolicy(
		ctx context.Context,
		meshService *v1alpha1.MeshService,
		mesh *v1alpha1.Mesh,
		mergedTrafficPolicy []*networking_v1alpha1.TrafficPolicy,
	) *types.TrafficPolicyStatus_TranslatorError
}

type TrafficPolicyTranslatorLoop interface {
	Start(ctx context.Context) error
}
