package internal

import (
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {

	MakeMeshServiceTranslator(
		meshes discoveryv1alpha2sets.MeshSet,
	) meshservice.Translator

}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshServiceTranslator(
	meshes discoveryv1alpha2sets.MeshSet,
) meshservice.Translator {
	return meshservice.NewTranslator(meshes)
}
