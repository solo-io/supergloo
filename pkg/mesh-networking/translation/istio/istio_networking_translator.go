package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/extensions"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/local"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	istioextensions "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/extensions"
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

	// note: these interfaces are set directly in unit tests, but not exposed in the Translator's constructor
	dependencies internal.DependencyFactory
	extensions   istioextensions.IstioExtender
}

func NewIstioTranslator(extensionClients extensions.Clientset) Translator {
	return &istioTranslator{
		dependencies: internal.NewDependencyFactory(),
		extensions:   istioextensions.NewIstioExtender(extensionClients),
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

	if err := t.extensions.PatchOutputs(ctx, in, istioOutputs); err != nil {
		// TODO(ilackarms): consider providing/checking user option to fail here when the extensions server is unavailable.
		// currently we just log the error and continue.
		contextutils.LoggerFrom(ctx).Errorf("failed to apply extension patches: %v", err)
	}

	t.totalTranslates++
}
