package translation

import (
	"context"

	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/mesh"
	meshdetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/mesh/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/mesh/detector/consul"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/mesh/detector/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/mesh/detector/linkerd"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/meshservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/meshworkload"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/meshworkload/detector"
	istiosidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/meshworkload/detector/istio"
	linkerdsidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshot/translation/meshworkload/detector/linkerd"
)

// the dependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our dependencyFactory anywhere else
type dependencyFactory interface {
	makeMeshTranslator(
		ctx context.Context,
		configMaps corev1sets.ConfigMapSet,
	) mesh.Translator

	makeMeshWorkloadTranslator(
		ctx context.Context,
		pods corev1sets.PodSet,
		replicaSets appsv1sets.ReplicaSetSet,
	) meshworkload.Translator

	makeMeshServiceTranslator() meshservice.Translator
}

type dependencyFactoryImpl struct{}

func (d dependencyFactoryImpl) makeMeshTranslator(ctx context.Context, configMaps corev1sets.ConfigMapSet) mesh.Translator {

	detectors := meshdetector.MeshDetectors{
		consul.NewMeshDetector(),
		istio.NewMeshDetector(configMaps),
		linkerd.NewMeshDetector(configMaps),
	}

	return mesh.NewTranslator(ctx, detectors)
}

func (d dependencyFactoryImpl) makeMeshWorkloadTranslator(
	ctx context.Context,
	pods corev1sets.PodSet,
	replicaSets appsv1sets.ReplicaSetSet,
) meshworkload.Translator {
	sidecarDetectors := detector.SidecarDetectors{
		istiosidecar.NewSidecarDetector(ctx),
		linkerdsidecar.NewSidecarDetector(ctx),
	}

	workloadDetector := detector.NewMeshWorkloadDetector(
		ctx,
		pods,
		replicaSets,
		sidecarDetectors,
	)
	return meshworkload.NewTranslator(workloadDetector)
}

func (d dependencyFactoryImpl) makeMeshServiceTranslator() meshservice.Translator {
	return meshservice.NewTranslator()

}
