package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/internal"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	// Translate translates the appropriate resources to apply input configuration resources for all Istio meshes contained in the input snapshot.
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		outputs output.Builder,
		reporter reporting.Reporter,
	)
}

type istioTranslator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    internal.DependencyFactory
}

func NewIstioTranslator() Translator {
	return &istioTranslator{
		dependencies: internal.NewDependencyFactory(),
	}
}

func (t *istioTranslator) Translate(
	ctx context.Context,
	in input.Snapshot,
	outputs output.Builder,
	reporter reporting.Reporter,
) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translator-%v", t.totalTranslates))

	meshServiceTranslator := t.dependencies.MakeMeshServiceTranslator(in.KubernetesClusters(), in.MeshServices())

	for _, meshService := range in.MeshServices().List() {
		meshService := meshService // pike

		meshServiceTranslator.Translate(in, meshService, outputs, reporter)
	}

	meshTranslator := t.dependencies.MakeMeshTranslator(ctx, in.KubernetesClusters(), in.Secrets(), in.MeshWorkloads(), in.MeshServices())
	for _, mesh := range in.Meshes().List() {
		meshTranslator.Translate(in, mesh, outputs, reporter)
	}

	t.totalTranslates++
}
