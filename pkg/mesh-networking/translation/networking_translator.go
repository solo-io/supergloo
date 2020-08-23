package translation

import (
	"bytes"
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	istiooutput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	smioutput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/smi"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/pkg/multicluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Outputs struct {
	Istio istiooutput.Snapshot
	Smi   smioutput.Snapshot
}

type TranslationResult struct {
	Istio istiooutput.Snapshot
	SMI   smioutput.Snapshot
}

func (t TranslationResult) MarshalJSON() ([]byte, error) {
	istioByt, err := t.Istio.MarshalJSON()
	if err != nil {
		return nil, err
	}
	smiByt, err := t.SMI.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return bytes.Join([][]byte{istioByt, smiByt}, []byte("\n")), nil
}

func (t TranslationResult) Apply(
	ctx context.Context,
	clusterClient client.Client,
	multiClusterClient multicluster.Client,
	errHandler output.ErrorHandler,
) {
	t.Istio.ApplyLocalCluster(ctx, clusterClient, errHandler)
	t.Istio.ApplyMultiCluster(ctx, multiClusterClient, errHandler)
	t.SMI.ApplyMultiCluster(ctx, multiClusterClient, errHandler)
}

// the networking translator translates an istio input networking snapshot to an istiooutput snapshot of mesh config resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		in input.Snapshot,
		reporter reporting.Reporter,
	) (TranslationResult, error)
}

type translator struct {
	totalTranslates int // TODO(ilackarms): metric
	istioTranslator istio.Translator
	smiTranslator   smi.Translator
	osmTranslator   osm.Translator
}

func NewTranslator(
	istioTranslator istio.Translator,
	smiTranslator smi.Translator,
	osmTranslator osm.Translator,
) Translator {
	return &translator{
		istioTranslator: istioTranslator,
		smiTranslator:   smiTranslator,
		osmTranslator:   osmTranslator,
	}
}

func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	reporter reporting.Reporter,
) (TranslationResult, error) {
	t.totalTranslates++
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("translation-%v", t.totalTranslates))

	istioOutputs := istiooutput.NewBuilder(ctx, fmt.Sprintf("networking-istio-%v", t.totalTranslates))
	smiOutputs := smioutput.NewBuilder(ctx, fmt.Sprintf("networking-smi-%v", t.totalTranslates))

	t.istioTranslator.Translate(ctx, in, istioOutputs, reporter)

	t.smiTranslator.Translate(ctx, in, smiOutputs, reporter)

	t.osmTranslator.Translate(ctx, in, smiOutputs, reporter)

	istioSnapshot, err := istioOutputs.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return TranslationResult{}, err
	}

	smiSnapshot, err := smiOutputs.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return TranslationResult{}, err
	}

	return TranslationResult{
		Istio: istioSnapshot,
		SMI:   smiSnapshot,
	}, nil
}
