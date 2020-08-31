package translation

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh"
	meshdetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/osm"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/traffictarget"
	traffictargetdetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/traffictarget/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload"
	workloaddetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector"
	istiosidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector/istio"
	osmsidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/detector/osm"
)

// we must generate in the same package because the interface is private
//go:generate mockgen -source ./dependencies.go -destination mocks/dependencies.go

// the dependencyFactory creates dependencies for the translator from a given snapshot
// NOTE(ilackarms): private interface used here as it's not expected we'll need to
// define our dependencyFactory anywhere else
type DependencyFactory interface {
	MakeMeshTranslator(
		ctx context.Context,
		in input.Snapshot,
	) mesh.Translator

	MakeWorkloadTranslator(
		ctx context.Context,
		in input.Snapshot,
	) workload.Translator

	MakeTrafficTargetTranslator(ctx context.Context) traffictarget.Translator
}

type DependencyFactoryImpl struct{}

func (d DependencyFactoryImpl) MakeMeshTranslator(ctx context.Context, in input.Snapshot) mesh.Translator {

	detectors := meshdetector.MeshDetectors{
		// TODO(EItanya): Uncomment to re-enable consul discovery
		// consul.NewMeshDetector(),
		istio.NewMeshDetector(
			ctx,
			in.ConfigMaps(),
			in.Services(),
			in.Pods(),
			in.Nodes(),
		),
		// TODO(EItanya): Uncomment to re-enable linkerd discovery
		// linkerd.NewMeshDetector(
		// 	in.ConfigMaps(),
		// ),
		osm.NewMeshDetector(ctx),
	}

	return mesh.NewTranslator(ctx, detectors)
}

func (d DependencyFactoryImpl) MakeWorkloadTranslator(
	ctx context.Context,
	in input.Snapshot,
) workload.Translator {
	sidecarDetectors := workloaddetector.SidecarDetectors{
		istiosidecar.NewSidecarDetector(ctx),
		// TODO(EItanya): Uncomment to re-enable linkerd discovery
		// linkerdsidecar.NewSidecarDetector(ctx),
		osmsidecar.NewSidecarDetector(ctx),
	}

	workloadDetector := workloaddetector.NewWorkloadDetector(
		ctx,
		in.Pods(),
		in.ReplicaSets(),
		sidecarDetectors,
	)
	return workload.NewTranslator(ctx, workloadDetector)
}

func (d DependencyFactoryImpl) MakeTrafficTargetTranslator(ctx context.Context) traffictarget.Translator {
	return traffictarget.NewTranslator(ctx, traffictargetdetector.NewTrafficTargetDetector(ctx))

}
