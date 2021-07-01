package decorators

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

//go:generate mockgen -source ./decorators.go -destination mocks/decorators.go

// parameters for initializing decorators
type Parameters struct {
	ClusterDomains hostutils.ClusterDomainRegistry
	Snapshot       input.LocalSnapshot
}

func Register(constructor Constructor) {
	registeredDecorators = append(registeredDecorators, constructor)
}

// Note: Translator decorators should be added here by the decorator in the init() function.
var registeredDecorators []Constructor

type Constructor func(params Parameters) Decorator

func makeDecorators(params Parameters) []Decorator {
	var decorators []Decorator
	for _, decoratorFactory := range registeredDecorators {
		decorator := decoratorFactory(params)
		decorators = append(decorators, decorator)
	}
	return decorators
}

// the decorator Factory initializes Translator decorators on each reconcile
type Factory interface {
	// return a set of decorators built from the given snapshot.
	MakeDecorators(params Parameters) []Decorator
}

type factory struct{}

func NewFactory() Factory {
	return &factory{}
}

func (f *factory) MakeDecorators(params Parameters) []Decorator {
	return makeDecorators(params)
}

// Decorators modify the output VirtualService corresponding to the input Destination.
type Decorator interface {
	// unique identifier for decorator
	DecoratorName() string
}

type RegisterField func(fieldPtr, val interface{}) error

/*
	Interface definitions for decorators which take VirtualMesh as an input and
	decorate a given output resource.
*/

// a VirtualDestinationEntryDecorator modifies the ServiceEntry based on a VirtualMesh which applies to the Destination.
type VirtualDestinationEntryDecorator interface {
	Decorator

	ApplyVirtualMeshToServiceEntry(
		appliedVirtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
		service *discoveryv1.Destination,
		output *networkingv1alpha3spec.ServiceEntry,
		registerField RegisterField,
	) error
}

/*
	Interface definitions for decorators which take TrafficPolicy as an input and
	decorate a given output resource.
*/

// a TrafficPolicyDestinationRuleDecorator modifies the DestinationRule based on a TrafficPolicy which applies to the Destination.
type TrafficPolicyDestinationRuleDecorator interface {
	Decorator

	ApplyTrafficPolicyToDestinationRule(
		appliedPolicy *networkingv1.AppliedTrafficPolicy,
		service *discoveryv1.Destination,
		output *networkingv1alpha3spec.DestinationRule,
		registerField RegisterField,
	) error
}

/*
	A TrafficPolicyVirtualServiceDecorator modifies the VirtualService based on a TrafficPolicy which applies to the Destination.

	If sourceMeshInstallation is specified, hostnames in the translated VirtualService will use global FQDNs if the Destination
	exists in a different cluster from the specified mesh (i.e. is a federated Destination). Otherwise, assume translation
	for cluster that the Destination exists in and use local FQDNs.
*/
type TrafficPolicyVirtualServiceDecorator interface {
	Decorator

	ApplyTrafficPolicyToVirtualService(
		appliedPolicy *networkingv1.AppliedTrafficPolicy,
		destination *discoveryv1.Destination,
		sourceMeshInstallation *discoveryv1.MeshInstallation,
		output *networkingv1alpha3spec.HTTPRoute,
		registerField RegisterField,
	) error
}
