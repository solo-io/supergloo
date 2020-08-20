package translation

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh"
	meshdetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/consul"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/linkerd"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/traffictarget"
	traffictargetdetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/traffictarget/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload"
	workloaddetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector"
	istiosidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector/istio"
	linkerdsidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector/linkerd"
)

// we must generate in the same package because the interface is private
//go:generate mockgen -source ./dependencies.go -destination mock_dependencies.go -package translation

// the dependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our dependencyFactory anywhere else
type dependencyFactory interface {
	makeMeshTranslator(
		ctx context.Context,
		in input.Snapshot,
	) mesh.Translator

	makeWorkloadTranslator(
		ctx context.Context,
		in input.Snapshot,
	) workload.Translator

	makeTrafficTargetTranslator(ctx context.Context) traffictarget.Translator
}

type dependencyFactoryImpl struct{}

func (d dependencyFactoryImpl) makeMeshTranslator(ctx context.Context, in input.Snapshot) mesh.Translator {

	detectors := meshdetector.MeshDetectors{
		consul.NewMeshDetector(),
		istio.NewMeshDetector(
			ctx,
			in.ConfigMaps(),
			in.Services(),
			in.Pods(),
			in.Nodes(),
		),
		linkerd.NewMeshDetector(
			in.ConfigMaps(),
		),
	}

	return mesh.NewTranslator(ctx, detectors)
}

func (d dependencyFactoryImpl) makeWorkloadTranslator(
	ctx context.Context,
	in input.Snapshot,
) workload.Translator {
	sidecarDetectors := workloaddetector.SidecarDetectors{
		istiosidecar.NewSidecarDetector(ctx),
		linkerdsidecar.NewSidecarDetector(ctx),
	}

	workloadDetector := workloaddetector.NewWorkloadDetector(
		ctx,
		in.Pods(),
		in.ReplicaSets(),
		sidecarDetectors,
	)
	return workload.NewTranslator(ctx, workloadDetector)
}

func (d dependencyFactoryImpl) makeTrafficTargetTranslator(ctx context.Context) traffictarget.Translator {
	return traffictarget.NewTranslator(ctx, traffictargetdetector.NewTrafficTargetDetector())

}
