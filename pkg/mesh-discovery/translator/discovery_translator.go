package translator

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/output"
	"github.com/solo-io/smh/pkg/mesh-discovery/utils/labelutils"
)

// the translator "reconciles the entire state of the world"
type Translator interface {
	// translates the Input Snapshot to an Output Snapshot
	Translate(ctx context.Context, in input.Snapshot) (output.Snapshot, error)
}

type translator struct {
	dependencies dependencyFactory
}

func NewTranslator() Translator {
	return &translator{dependencies: dependencyFactoryImpl{}}
}

func (t translator) Translate(ctx context.Context, in input.Snapshot) (output.Snapshot, error) {

	meshTranslator := t.dependencies.makeMeshTranslator(ctx,
		in.ConfigMaps(),
	)

	meshWorkloadTranslator := t.dependencies.makeMeshWorkloadTranslator(ctx,
		in.Pods(),
		in.ReplicaSets(),
	)

	meshServiceTranslator := t.dependencies.makeMeshServiceTranslator()

	meshes := meshTranslator.TranslateMeshes(in.Deployments())

	meshWorkloads := meshWorkloadTranslator.TranslateMeshWorkloads(
		in.Deployments(),
		in.DaemonSets(),
		in.StatefulSets(),
		meshes,
	)

	meshServices := meshServiceTranslator.TranslateMeshServices(in.Services(), meshWorkloads)

	return output.NewLabelPartitionedSnapshot(
		"mesh-discovery",
		labelutils.ClusterLabelKey,
		meshServices,
		meshWorkloads,
		meshes,
	)
}
