package istio

import (
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	"github.com/solo-io/smh/pkg/mesh-networking/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
)

// the dependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our dependencyFactory anywhere else
type dependencyFactory interface {
	makeMeshServiceTranslator(clusters skv1alpha1sets.KubernetesClusterSet) meshservice.Translator
}

type dependencyFactoryImpl struct{}

func newDependencyFactory() dependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) makeMeshServiceTranslator(clusters skv1alpha1sets.KubernetesClusterSet) meshservice.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	decoratorFactory := decorators.NewFactory()

	return meshservice.NewTranslator(clusterDomains, decoratorFactory)
}
