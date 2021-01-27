package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/output/discovery"
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
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
		settings *settingsv1alpha2.DiscoverySettings,
		localSnapshot input.SettingsSnapshot,
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
	settings *settingsv1alpha2.DiscoverySettings,
	localSnapshot input.SettingsSnapshot,
) (discovery.Snapshot, error) {

	meshTranslator := t.dependencies.MakeMeshTranslator(ctx)

	workloadTranslator := t.dependencies.MakeWorkloadTranslator(ctx, in)

	trafficTargetTranslator := t.dependencies.MakeTrafficTargetTranslator()

	meshes := meshTranslator.TranslateMeshes(in, settings)

	workloads := workloadTranslator.TranslateWorkloads(
		in.Appsv1Deployments(),
		in.Appsv1DaemonSets(),
		in.Appsv1StatefulSets(),
		meshes,
	)

	trafficTargets := trafficTargetTranslator.TranslateTrafficTargets(
		ctx,
		in.V1Services(),
		in.V1Endpoints(),
		workloads,
		meshes,
		localSnapshot.NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes(),
	)

	t.totalTranslates++

	return discovery.NewSinglePartitionedSnapshot(
		fmt.Sprintf("mesh-discovery-%v", t.totalTranslates),
		labelutils.OwnershipLabels(),
		trafficTargets,
		workloads,
		meshes,
	)
}
