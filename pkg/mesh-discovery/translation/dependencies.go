package translation

import (
	"context"

	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh"
	meshdetector "github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh/detector/consul"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh/detector/istio"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh/detector/linkerd"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/meshservice"
	meshservicedetector "github.com/solo-io/smh/pkg/mesh-discovery/translation/meshservice/detector"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/meshworkload"
	meshworkloaddetector "github.com/solo-io/smh/pkg/mesh-discovery/translation/meshworkload/detector"
	istiosidecar "github.com/solo-io/smh/pkg/mesh-discovery/translation/meshworkload/detector/istio"
	linkerdsidecar "github.com/solo-io/smh/pkg/mesh-discovery/translation/meshworkload/detector/linkerd"
)

// we must generate in the same package because the interface is private
//go:generate mockgen -source ./dependencies.go -destination mock_dependencies.go -package translation

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

	makeMeshServiceTranslator(ctx context.Context) meshservice.Translator
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
	sidecarDetectors := meshworkloaddetector.SidecarDetectors{
		istiosidecar.NewSidecarDetector(ctx),
		linkerdsidecar.NewSidecarDetector(ctx),
	}

	workloadDetector := meshworkloaddetector.NewMeshWorkloadDetector(
		ctx,
		pods,
		replicaSets,
		sidecarDetectors,
	)
	return meshworkload.NewTranslator(ctx, workloadDetector)
}

func (d dependencyFactoryImpl) makeMeshServiceTranslator(ctx context.Context) meshservice.Translator {
	return meshservice.NewTranslator(ctx, meshservicedetector.NewMeshServiceDetector())

}
