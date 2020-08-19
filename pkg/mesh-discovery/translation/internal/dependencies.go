package translator_internal

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh"
	meshdetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector/osm"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshservice"
	meshservicedetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshservice/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshworkload"
	meshworkloaddetector "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshworkload/detector"
	istiosidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshworkload/detector/istio"
	osmsidecar "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshworkload/detector/osm"
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

	MakeMeshWorkloadTranslator(
		ctx context.Context,
		in input.Snapshot,
	) meshworkload.Translator

	MakeMeshServiceTranslator(ctx context.Context) meshservice.Translator
}

type DependencyFactoryImpl struct{}

func (d DependencyFactoryImpl) MakeMeshTranslator(ctx context.Context, in input.Snapshot) mesh.Translator {

	detectors := meshdetector.MeshDetectors{
		// TODO: Uncomment to re-enable consul discovery
		// consul.NewMeshDetector(),
		istio.NewMeshDetector(
			ctx,
			in.ConfigMaps(),
			in.Services(),
			in.Pods(),
			in.Nodes(),
		),
		// TODO: Uncomment to re-enable linkerd discovery
		// linkerd.NewMeshDetector(
		// 	in.ConfigMaps(),
		// ),
		osm.NewMeshDetector(ctx),
	}

	return mesh.NewTranslator(ctx, detectors)
}

func (d DependencyFactoryImpl) MakeMeshWorkloadTranslator(
	ctx context.Context,
	in input.Snapshot,
) meshworkload.Translator {
	sidecarDetectors := meshworkloaddetector.SidecarDetectors{
		istiosidecar.NewSidecarDetector(ctx),
		// TODO: Uncomment to re-enable linkerd discovery
		// linkerdsidecar.NewSidecarDetector(ctx),
		osmsidecar.NewSidecarDetector(ctx),
	}

	workloadDetector := meshworkloaddetector.NewMeshWorkloadDetector(
		ctx,
		in.Pods(),
		in.ReplicaSets(),
		sidecarDetectors,
	)
	return meshworkload.NewTranslator(ctx, workloadDetector)
}

func (d DependencyFactoryImpl) MakeMeshServiceTranslator(ctx context.Context) meshservice.Translator {
	return meshservice.NewTranslator(ctx, meshservicedetector.NewMeshServiceDetector())

}
