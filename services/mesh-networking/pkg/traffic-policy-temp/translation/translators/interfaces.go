package mesh_translation

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type IstioTranslator interface {
	TranslationValidator
	DiscoveryLabelsGetter

	Translate(
		meshService *zephyr_discovery.MeshService,
		mesh *zephyr_discovery.Mesh,
		trafficPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	) (*IstioTranslationOutput, []*TranslationError)
}

type TranslationValidator interface {
	GetTranslationErrors(
		meshService *zephyr_discovery.MeshService,
		mesh *zephyr_discovery.Mesh,
		trafficPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	) []*TranslationError
}

type DiscoveryLabelsGetter interface {
	// get the labels that this translator attaches to the resources it creates
	GetTranslationLabels() map[string]string
}

type TranslationError struct {
	Policy           *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
	TranslatorErrors []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError
}

type IstioTranslationOutput struct {
	VirtualServices  []*istio_networking.VirtualService
	DestinationRules []*istio_networking.DestinationRule
}
