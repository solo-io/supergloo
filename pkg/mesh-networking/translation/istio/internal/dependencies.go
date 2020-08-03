package internal

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/mtls"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/access"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/failoverservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeMeshServiceTranslator(clusters skv1alpha1sets.KubernetesClusterSet) meshservice.Translator
	MakeMeshTranslator(ctx context.Context, clusters skv1alpha1sets.KubernetesClusterSet) mesh.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshServiceTranslator(clusters skv1alpha1sets.KubernetesClusterSet) meshservice.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	decoratorFactory := decorators.NewFactory()

	return meshservice.NewTranslator(clusterDomains, decoratorFactory)
}

func (d dependencyFactoryImpl) MakeMeshTranslator(ctx context.Context, clusters skv1alpha1sets.KubernetesClusterSet) mesh.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	federationTranslator := federation.NewTranslator(ctx, clusterDomains)
	mtlsTranslator := mtls.NewTranslator(ctx)
	accessTranslator := access.NewTranslator()
	failoverServiceTranslator := failoverservice.NewTranslator(ctx, clusterDomains)

	return mesh.NewTranslator(
		ctx,
		mtlsTranslator,
		federationTranslator,
		accessTranslator,
		failoverServiceTranslator,
	)
}
