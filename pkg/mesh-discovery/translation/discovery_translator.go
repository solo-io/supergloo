package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/output/discovery"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	internal "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/internal"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/labelutils"
)

var DefaultDependencyFactory = internal.DependencyFactoryImpl{}

// the translator "reconciles the entire state of the world"
type Translator interface {
	// translates the Input Snapshot to an Output Snapshot
	Translate(
		ctx context.Context,
		in input.DiscoveryInputSnapshot,
		settings *settingsv1.DiscoverySettings,
	) (discovery.Snapshot, error)
}

type translator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    internal.DependencyFactory
}

func NewTranslator(dependencyFactory internal.DependencyFactory) Translator {
	return &translator{
		dependencies: dependencyFactory,
	}
}

func (t translator) Translate(
	ctx context.Context,
	in input.DiscoveryInputSnapshot,
	settings *settingsv1.DiscoverySettings,
) (discovery.Snapshot, error) {

	meshTranslator := t.dependencies.MakeMeshTranslator(ctx)

	workloadTranslator := t.dependencies.MakeWorkloadTranslator(ctx, in)

	destinationTranslator := t.dependencies.MakeDestinationTranslator()

	meshes := meshTranslator.TranslateMeshes(in, settings)

	workloads := workloadTranslator.TranslateWorkloads(
		in.Deployments(),
		in.DaemonSets(),
		in.StatefulSets(),
		meshes,
	)

	destinations := destinationTranslator.TranslateDestinations(
		ctx,
		in.Services(),
		in.Pods(),
		in.Nodes(),
		workloads,
		meshes,
		in.Endpoints(),
	)

	t.totalTranslates++

	return discovery.NewSinglePartitionedSnapshot(
		fmt.Sprintf("mesh-discovery-%v", t.totalTranslates),
		labelutils.OwnershipLabels(),
		destinations,
		workloads,
		meshes,
	)
}
