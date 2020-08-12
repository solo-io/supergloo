package internal

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"

	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/mtls"

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
	MakeMeshServiceTranslator(
		ctx context.Context,
		clusters skv1alpha1sets.KubernetesClusterSet,
		meshes discoveryv1alpha2sets.MeshSet,
		meshServices discoveryv1alpha2sets.MeshServiceSet,
	) meshservice.Translator

	MakeMeshTranslator(
		ctx context.Context,
		clusters skv1alpha1sets.KubernetesClusterSet,
		secrets corev1sets.SecretSet,
		meshWorkloads discoveryv1alpha2sets.MeshWorkloadSet,
		meshServices discoveryv1alpha2sets.MeshServiceSet,
	) mesh.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeMeshServiceTranslator(
	ctx context.Context,
	clusters skv1alpha1sets.KubernetesClusterSet,
	meshes discoveryv1alpha2sets.MeshSet,
	meshServices discoveryv1alpha2sets.MeshServiceSet,
) meshservice.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	decoratorFactory := decorators.NewFactory()

	return meshservice.NewTranslator(ctx, meshes, clusterDomains, decoratorFactory, meshServices)
}

func (d dependencyFactoryImpl) MakeMeshTranslator(
	ctx context.Context,
	clusters skv1alpha1sets.KubernetesClusterSet,
	secrets corev1sets.SecretSet,
	meshWorkloads discoveryv1alpha2sets.MeshWorkloadSet,
	meshServices discoveryv1alpha2sets.MeshServiceSet,
) mesh.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	federationTranslator := federation.NewTranslator(ctx, clusterDomains, meshServices)
	mtlsTranslator := mtls.NewTranslator(ctx, secrets, meshWorkloads)
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
