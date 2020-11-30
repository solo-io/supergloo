package internal

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input/user"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/destinationrule"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/virtualservice"

	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/access"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/failoverservice"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
)

//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the DependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our DependencyFactory anywhere else
type DependencyFactory interface {
	MakeTrafficTargetTranslator(
		ctx context.Context,
		userInputSnap user.Snapshot,
		clusters skv1alpha1sets.KubernetesClusterSet,
		trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
		failoverServices v1alpha2sets.FailoverServiceSet,
	) traffictarget.Translator
	MakeMeshTranslator(
		ctx context.Context,
		userInputSnap user.Snapshot,
		clusters skv1alpha1sets.KubernetesClusterSet,
		secrets corev1sets.SecretSet,
		workloads discoveryv1alpha2sets.WorkloadSet,
		trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
		failoverServices v1alpha2sets.FailoverServiceSet,
	) mesh.Translator
}

type dependencyFactoryImpl struct{}

func NewDependencyFactory() DependencyFactory {
	return dependencyFactoryImpl{}
}

func (d dependencyFactoryImpl) MakeTrafficTargetTranslator(
	ctx context.Context,
	userInputSnap user.Snapshot,
	clusters skv1alpha1sets.KubernetesClusterSet,
	trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	failoverServices v1alpha2sets.FailoverServiceSet,
) traffictarget.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	decoratorFactory := decorators.NewFactory()

	return traffictarget.NewTranslator(ctx, userInputSnap, clusterDomains, decoratorFactory, trafficTargets, failoverServices)
}

func (d dependencyFactoryImpl) MakeMeshTranslator(
	ctx context.Context,
	userInputSnap user.Snapshot,
	clusters skv1alpha1sets.KubernetesClusterSet,
	secrets corev1sets.SecretSet,
	workloads discoveryv1alpha2sets.WorkloadSet,
	trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	failoverServices v1alpha2sets.FailoverServiceSet,
) mesh.Translator {
	clusterDomains := hostutils.NewClusterDomainRegistry(clusters)
	federationTranslator := federation.NewTranslator(
		ctx,
		clusterDomains,
		trafficTargets,
		failoverServices,
		virtualservice.NewTranslator(userInputSnap, clusterDomains, decorators.NewFactory()),
		destinationrule.NewTranslator(userInputSnap, clusterDomains, decorators.NewFactory(), trafficTargets, failoverServices),
	)
	mtlsTranslator := mtls.NewTranslator(ctx, secrets, workloads)
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
