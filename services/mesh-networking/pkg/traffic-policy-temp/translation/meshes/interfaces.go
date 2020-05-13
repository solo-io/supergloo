package mesh_translation

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	istio_networking "istio.io/api/networking/v1alpha3"
)

type IstioTranslator interface {
	Translate(
		ctx context.Context,
		meshService *zephyr_discovery.MeshService,
		mesh *zephyr_discovery.Mesh,
		trafficPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	) (IstioTranslationOutput, []*TranslationError)
}

type TranslationValidator interface {
	GetTranslationErrors(ctx context.Context,
		meshService *zephyr_discovery.MeshService,
		mesh *zephyr_discovery.Mesh,
		trafficPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	) []*TranslationError
}

type TranslationError struct {
	Policy           *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
	TranslatorErrors []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError
}

type IstioTranslationOutput struct {
	VirtualServices  []*istio_networking.VirtualService
	DestinationRules []*istio_networking.DestinationRule
}
