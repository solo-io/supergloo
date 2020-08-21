package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	istioinput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	istiooutput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/output"
	smiinput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/smi/input"
	smioutput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/smi/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
)

type Inputs struct {
	Istio istioinput.Snapshot
	Smi   smiinput.Snapshot
}

type Outputs struct {
	Istio istiooutput.Snapshot
	Smi   smioutput.Snapshot
}

// the networking translator translates an istio input networking snapshot to an istiooutput snapshot of mesh config resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		in Inputs,
		reporter reporting.Reporter,
	) (Outputs, error)
}

type translator struct {
	totalTranslates int // TODO(ilackarms): metric
	istioTranslator istio.Translator
	smiTranslator   smi.Translator
}

func NewTranslator(
	istioTranslator istio.Translator,
	smiTranslator smi.Translator,
) Translator {
	return &translator{
		istioTranslator: istioTranslator,
		smiTranslator:   smiTranslator,
	}
}

func (t *translator) Translate(
	ctx context.Context,
	in Inputs,
	reporter reporting.Reporter,
) (Outputs, error) {
	t.totalTranslates++
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("translation-%v", t.totalTranslates))

	istioOutputs := istiooutput.NewBuilder(ctx, fmt.Sprintf("networking-%v", t.totalTranslates))

	t.istioTranslator.Translate(ctx, in.Istio, istioOutputs, reporter)

	t.smiTranslator.Translate(ctx, in, istioOutputs, reporter)

	buildIstioSnapshot, err := istioOutputs.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return Outputs{}, err
	}

	return Outputs{
		Istio: buildIstioSnapshot,
	}, nil
}
