package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
)

// the networking translator translates an input networking snapshot to an output snapshot of mesh config resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		in input.Snapshot,
		reporter reporting.Reporter,
	) (output.Snapshot, error)
}

type translator struct {
	totalTranslates int // TODO(ilackarms): metric
	istioTranslator istio.Translator
}

func NewTranslator(
	istioTranslator istio.Translator,
) Translator {
	return &translator{
		istioTranslator: istioTranslator,
	}
}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	reporter reporting.Reporter,
) (output.Snapshot, error) {
	t.totalTranslates++
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("translation-%v", t.totalTranslates))

	outputs := output.NewBuilder(ctx, fmt.Sprintf("networking-%v", t.totalTranslates))

	t.istioTranslator.Translate(ctx, in, outputs, reporter)

	return outputs.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
}
