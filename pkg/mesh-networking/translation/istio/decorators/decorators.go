package decorators

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

//go:generate mockgen -source ./decorators.go -destination mocks/decorators.go

// parameters for initializing decorators
type Parameters struct {
	ClusterDomains hostutils.ClusterDomainRegistry
	Snapshot       input.Snapshot
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

// Decorators modify the output VirtualService corresponding to the input TrafficTarget.
type Decorator interface {
	// unique identifier for decorator
	DecoratorName() string
}

type RegisterField func(fieldPtr, val interface{}) error

/*
	Interface definitions for decorators which take VirtualMesh as an input and
	decorate a given output resource.
*/

// a VirtualMeshDestinationRuleDecorator modifies the DestinationRule based on a VirtualMesh which applies to the TrafficTarget.
type VirtualMeshDestinationRuleDecorator interface {
	Decorator

	ApplyVirtualMeshToDestinationRule(
		appliedVirtualMesh *v1alpha2.MeshStatus_AppliedVirtualMesh,
		service *v1alpha2.TrafficTarget,
		output *networkingv1alpha3spec.DestinationRule,
		registerField RegisterField,
	) error
}

// a VirtualMeshServiceEntryDecorator modifies the ServiceEntry based on a VirtualMesh which applies to the TrafficTarget.
type VirtualMeshServiceEntryDecorator interface {
	Decorator

	ApplyVirtualMeshToServiceEntry(
		appliedVirtualMesh *v1alpha2.MeshStatus_AppliedVirtualMesh,
		service *v1alpha2.TrafficTarget,
		output *networkingv1alpha3spec.ServiceEntry,
		registerField RegisterField,
	) error
}

/*
	Interface definitions for decorators which take TrafficPolicy as an input and
	decorate a given output resource.
*/

// a TrafficPolicyDestinationRuleDecorator modifies the DestinationRule based on a TrafficPolicy which applies to the TrafficTarget.
type TrafficPolicyDestinationRuleDecorator interface {
	Decorator

	ApplyTrafficPolicyToDestinationRule(
		appliedPolicy *v1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
		service *v1alpha2.TrafficTarget,
		output *networkingv1alpha3spec.DestinationRule,
		registerField RegisterField,
	) error
}

// a TrafficPolicyVirtualServiceDecorator modifies the VirtualService based on a TrafficPolicy which applies to the TrafficTarget.
type TrafficPolicyVirtualServiceDecorator interface {
	Decorator

	ApplyTrafficPolicyToVirtualService(
		appliedPolicy *v1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
		service *v1alpha2.TrafficTarget,
		output *networkingv1alpha3spec.HTTPRoute,
		registerField RegisterField,
	) error
}
