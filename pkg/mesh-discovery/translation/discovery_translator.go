package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/output"
	translator_internal "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/internal"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
)

// the translator "reconciles the entire state of the world"
type Translator interface {
	// translates the Input Snapshot to an Output Snapshot
	Translate(ctx context.Context, in input.Snapshot) (output.Snapshot, error)
}

type translator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    translator_internal.DependencyFactory
}

func NewTranslator() Translator {
	return &translator{
		dependencies: translator_internal.DependencyFactoryImpl{},
	}
}

func (t translator) Translate(ctx context.Context, in input.Snapshot) (output.Snapshot, error) {

	meshTranslator := t.dependencies.MakeMeshTranslator(ctx, in)

	meshWorkloadTranslator := t.dependencies.MakeMeshWorkloadTranslator(ctx, in)

	meshServiceTranslator := t.dependencies.MakeMeshServiceTranslator(ctx)

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
