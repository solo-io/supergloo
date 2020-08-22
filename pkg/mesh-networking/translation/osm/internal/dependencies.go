package internal

import (
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm/mesh"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeMeshTranslator() mesh.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshTranslator() mesh.Translator {
	return mesh.NewTranslator()
}
