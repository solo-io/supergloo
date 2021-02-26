package internal

import (
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm/destination"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm/mesh"
	smitraffictarget "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination/access"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination/split"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeMeshTranslator() mesh.Translator
	MakeDestinationTranslator() destination.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshTranslator() mesh.Translator {
	return mesh.NewTranslator()
}

func (d dependencyFactoryImpl) MakeDestinationTranslator() destination.Translator {
	splitTranslator := split.NewTranslator()
	accessTranslator := access.NewTranslator()
	trafficTargetTranslator := smitraffictarget.NewTranslator(splitTranslator, accessTranslator)
	return destination.NewTranslator(trafficTargetTranslator)
}
