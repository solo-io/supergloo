package internal

import (
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm/mesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm/traffictarget"
	smitraffictarget "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/access"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/split"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeMeshTranslator() mesh.Translator
	MakeTrafficTargetTranslator() traffictarget.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshTranslator() mesh.Translator {
	return mesh.NewTranslator()
}

func (d dependencyFactoryImpl) MakeTrafficTargetTranslator() traffictarget.Translator {
	splitTranslator := split.NewTranslator()
	accessTranslator := access.NewTranslator()
	trafficTargetTranslator := smitraffictarget.NewTranslator(splitTranslator, accessTranslator)
	return traffictarget.NewTranslator(trafficTargetTranslator)
}
