package internal

import (
	"context"

	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeMeshServiceTranslator(
		ctx context.Context,
		clusters skv1alpha1sets.KubernetesClusterSet,
		meshes discoveryv1alpha2sets.MeshSet,
	) meshservice.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshServiceTranslator(
	ctx context.Context,
	clusters skv1alpha1sets.KubernetesClusterSet,
	meshes discoveryv1alpha2sets.MeshSet,
) meshservice.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	decoratorFactory := decorators.NewFactory()

	return meshservice.NewTranslator(ctx, meshes, clusterDomains, decoratorFactory)
}
