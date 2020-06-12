package mesh_translation

import (
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

/*
 * Every translator should embed the following three interfaces, for consistency across all the different mesh-specific translators.
 * Note that the validator is used in the aggregation validation step.
 */
type TranslationValidator interface {
	// Ignore the mesh-specific output from the translation step, in an effort to unify them behind some kind of common interface.
	// Generics ;(
	GetTranslationErrors(
		meshService *smh_discovery.MeshService,
		allMeshServices []*smh_discovery.MeshService,
		mesh *smh_discovery.Mesh,
		trafficPolicies []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	) []*TranslationError
}

type DiscoveryLabelsGetter interface {
	// get the labels that this translator attaches to the resources it creates
	GetTranslationLabels() map[string]string
}

type NamedTranslator interface {
	Name() string
}

// Mesh-specific interfaces

type IstioTranslator interface {
	TranslationValidator
	DiscoveryLabelsGetter
	NamedTranslator
	snapshot.TranslationSnapshotAccumulator
	Translate(
		meshService *smh_discovery.MeshService,
		allMeshServices []*smh_discovery.MeshService,
		mesh *smh_discovery.Mesh,
		trafficPolicies []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	) (*IstioTranslationOutput, []*TranslationError)
}

type TranslationError struct {
	Policy           *smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy
	TranslatorErrors []*smh_networking_types.TrafficPolicyStatus_TranslatorError
}

type IstioTranslationOutput struct {
	VirtualServices  []*istio_networking.VirtualService
	DestinationRules []*istio_networking.DestinationRule
}
