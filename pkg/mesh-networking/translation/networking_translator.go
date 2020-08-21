package translation

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	istiooutput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	smioutput "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/smi"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
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

func (r TranslationResult) ApplyLocalCluster(ctx context.Context, clusterClient client.Client, errHandler output.ErrorHandler) {
	r.Istio.ApplyLocalCluster(ctx, clusterClient, errHandler)
	r.SMI.ApplyLocalCluster(ctx, clusterClient, errHandler)
}

func (r TranslationResult) ApplyMultiCluster(ctx context.Context, multiClusterClient multicluster.Client, errHandler output.ErrorHandler) {
	r.Istio.ApplyMultiCluster(ctx, multiClusterClient, errHandler)
	r.SMI.ApplyMultiCluster(ctx, multiClusterClient, errHandler)

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
	in input.Snapshot,
	reporter reporting.Reporter,
) (TranslationResult, error) {
	t.totalTranslates++
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("translation-%v", t.totalTranslates))

	istioOutputs := istiooutput.NewBuilder(ctx, fmt.Sprintf("networking-istio-%v", t.totalTranslates))
	smiOutputs := smioutput.NewBuilder(ctx, fmt.Sprintf("networking-smi-%v", t.totalTranslates))

	t.istioTranslator.Translate(ctx, in, istioOutputs, reporter)

	t.smiTranslator.Translate(ctx, in, smiOutputs, reporter)

	buildIstioSnapshot, err := istioOutputs.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return TranslationResult{}, err
	}

	return TranslationResult{
		Istio: buildIstioSnapshot,
	}, nil
}
