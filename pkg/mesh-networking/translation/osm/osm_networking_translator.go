package osm

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/smi"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm/internal"
	"github.com/solo-io/go-utils/contextutils"
)

var DefaultDependencyFactory = internal.NewDependencyFactory()

//go:generate mockgen -source ./osm_networking_translator.go -destination mocks/osm_networking_translator.go

// the smi translator translates an input networking snapshot to an output snapshot of SMI resources
type Translator interface {
	// Translate translates the appropriate resources to apply input configuration resources for all OSM meshes contained in the input snapshot.
	// Output resources will be added to the smi.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
		outputs smi.Builder,
		reporter reporting.Reporter,
	)
}

type osmTranslator struct {
	totalTranslates int // TODO(ilackarms): metric
	dependencies    internal.DependencyFactory
}

func NewOSMTranslator() Translator {
	return &osmTranslator{
		dependencies: internal.NewDependencyFactory(),
	}
}

func (s *osmTranslator) Translate(
	ctx context.Context,
	in input.LocalSnapshot,
	outputs smi.Builder,
	reporter reporting.Reporter,
) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("osm-translator-%v", s.totalTranslates))

	meshTranslator := s.dependencies.MakeMeshTranslator()

	for _, mesh := range in.Meshes().List() {
		mesh := mesh

		meshTranslator.Translate(ctx, in, mesh, outputs, reporter)
	}

	destinationTranslator := s.dependencies.MakeDestinationTranslator()

	for _, destination := range in.Destinations().List() {
		destination := destination

		destinationTranslator.Translate(ctx, in, destination, outputs, reporter)
	}

	s.totalTranslates++
}
