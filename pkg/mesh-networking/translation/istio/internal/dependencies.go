package internal

import (
	"context"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/access"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeDestinationTranslator(
		ctx context.Context,
		userSupplied input.RemoteSnapshot,
		clusters skv1alpha1sets.KubernetesClusterSet,
		destinations discoveryv1sets.DestinationSet,
	) destination.Translator
	MakeMeshTranslator(
		ctx context.Context,
		secrets corev1sets.SecretSet,
		workloads discoveryv1sets.WorkloadSet,
	) mesh.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeDestinationTranslator(
	ctx context.Context,
	userSupplied input.RemoteSnapshot,
	clusters skv1alpha1sets.KubernetesClusterSet,
	destinations discoveryv1sets.DestinationSet,
) destination.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters, destinations)
	decoratorFactory := decorators.NewFactory()

	return destination.NewTranslator(ctx, userSupplied, clusterDomains, decoratorFactory, destinations)
}

func (d dependencyFactoryImpl) MakeMeshTranslator(
	ctx context.Context,
	secrets corev1sets.SecretSet,
	workloads discoveryv1sets.WorkloadSet,
) mesh.Translator {
	federationTranslator := federation.NewTranslator(ctx)
	mtlsTranslator := mtls.NewTranslator(ctx, secrets, workloads)
	accessTranslator := access.NewTranslator(ctx)

	return mesh.NewTranslator(
		ctx,
		mtlsTranslator,
		federationTranslator,
		accessTranslator,
	)
}
