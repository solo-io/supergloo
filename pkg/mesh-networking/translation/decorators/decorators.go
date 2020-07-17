package decorators

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
)

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

// Decorators modify the output VirtualService corresponding to the input MeshService.
type Decorator interface {
	// unique identifier for decorator
	DecoratorName() string
}

type RegisterField func(fieldPtr, val interface{}) error
