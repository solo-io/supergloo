package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
)

// the translator "reconciles the entire state of the world"
type Translator interface {
	// translates the Input Snapshot to an Output Snapshot
	Translate(ctx context.Context, in input.Snapshot) (output.Snapshot, error)
}

type translator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    dependencyFactory
}

func NewTranslator() Translator {
	return &translator{
		dependencies: dependencyFactoryImpl{},
	}
}

func (t translator) Translate(ctx context.Context, in input.Snapshot) (output.Snapshot, error) {

	meshTranslator := t.dependencies.makeMeshTranslator(ctx, in)

	meshWorkloadTranslator := t.dependencies.makeMeshWorkloadTranslator(ctx, in)

	meshServiceTranslator := t.dependencies.makeMeshServiceTranslator(ctx)

	meshes := meshTranslator.TranslateMeshes(in.Deployments())

	meshWorkloads := meshWorkloadTranslator.TranslateMeshWorkloads(
		in.Deployments(),
		in.DaemonSets(),
		in.StatefulSets(),
		meshes,
	)

	meshServices := meshServiceTranslator.TranslateMeshServices(in.Services(), meshWorkloads)

	t.totalTranslates++

	return output.NewSinglePartitionedSnapshot(
		fmt.Sprintf("mesh-discovery-%v", t.totalTranslates),
		labelutils.OwnershipLabels(),
		meshServices,
		meshWorkloads,
		meshes,
	)
}
