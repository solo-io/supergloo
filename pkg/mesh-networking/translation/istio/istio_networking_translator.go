package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	Translate(
		ctx context.Context,
		in input.Snapshot,
		outputs output.Builder,
		reporter reporting.Reporter,
	)
}

type istioTranslator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    dependencyFactory
}

func NewIstioTranslator() Translator {
	return &istioTranslator{
		dependencies: dependencyFactoryImpl{},
	}
}

func (t *istioTranslator) Translate(
	ctx context.Context,
	in input.Snapshot,
	outputs output.Builder,
	reporter reporting.Reporter,
) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translator-%v", t.totalTranslates))

	meshServiceTranslator := t.dependencies.makeMeshServiceTranslator(in.KubernetesClusters())

	for _, meshService := range in.MeshServices().List() {
		meshService := meshService // pike

		meshServiceTranslator.Translate(in, meshService, outputs, reporter)
	}

	meshTranslator := t.dependencies.makeMeshTranslator(ctx, in.KubernetesClusters())
	for _, mesh := range in.Meshes().List() {
		meshTranslator.Translate(in, mesh, outputs, reporter)
	}

	t.totalTranslates++
}
