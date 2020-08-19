package smi

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/internal"
)

var DefaultDependencyFactory = internal.NewDependencyFactory()

// the smi translator translates an input networking snapshot to an output snapshot of SMI resources
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

type smiTranslator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    internal.DependencyFactory
}

func NewSmiTranslator(deps internal.DependencyFactory) Translator {
	return &smiTranslator{
		dependencies: deps,
	}
}

func (s *smiTranslator) Translate(
	ctx context.Context,
	in input.Snapshot,
	outputs output.Builder,
	reporter reporting.Reporter,
) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("smi-translator-%v", s.totalTranslates))

	meshServiceTranslator := s.dependencies.MakeMeshServiceTranslator()

	for _, meshService := range in.MeshServices().List() {
		meshService := meshService // pike

		meshServiceTranslator.Translate(ctx, in, meshService, outputs, reporter)
	}

	s.totalTranslates++
}
