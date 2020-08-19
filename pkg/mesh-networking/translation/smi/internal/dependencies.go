package internal

import (
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice/access"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice/split"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeMeshServiceTranslator() meshservice.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshServiceTranslator() meshservice.Translator {
	return meshservice.NewTranslator(split.NewTrafficSplitTranslator(), access.NewTranslator())
}
