package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/local"
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
		istioOutputs istio.Builder,
		localOutputs local.Builder,
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
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	reporter reporting.Reporter,
) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translator-%v", t.totalTranslates))

	trafficTargetTranslator := t.dependencies.MakeTrafficTargetTranslator(
		ctx,
		in.KubernetesClusters(),
		in.TrafficTargets(),
		in.FailoverServices(),
	)

	for _, trafficTarget := range in.TrafficTargets().List() {
		trafficTarget := trafficTarget // pike

		trafficTargetTranslator.Translate(in, trafficTarget, istioOutputs, reporter)
	}

	meshTranslator := t.dependencies.MakeMeshTranslator(
		ctx,
		in.KubernetesClusters(),
		in.Secrets(),
		in.Workloads(),
		in.TrafficTargets(),
		in.FailoverServices(),
	)

	for _, mesh := range in.Meshes().List() {
		meshTranslator.Translate(in, mesh, istioOutputs, localOutputs, reporter)
	}

	t.totalTranslates++
}
