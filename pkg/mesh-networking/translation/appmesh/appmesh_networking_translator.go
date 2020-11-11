package appmesh

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/appmesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/go-utils/contextutils"
)

// the appmesh translator translates an input networking snapshot to an output snapshot of appmesh resources
type Translator interface {
	// Translate translates the appropriate resources to apply input configuration resources for all appmesh meshes contained in the input snapshot.
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		appmeshOutputs appmesh.Builder,
		reporter reporting.Reporter,
	)
}

type appmeshTranslator struct {
	totalTranslates int // TODO(ilackarms): metric
}

func NewAppmeshTranslator() Translator {
	return &appmeshTranslator{}
}

func (t *appmeshTranslator) Translate(
	ctx context.Context,
	in input.Snapshot,
	appmeshOutputs appmesh.Builder,
	reporter reporting.Reporter,
) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("appmesh-translator-%v", t.totalTranslates))

	//TODO: implement AppMesh translation

	t.totalTranslates++
}
