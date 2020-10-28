package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/extensions"

	istioextensions "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/extensions"
	"github.com/solo-io/skv2/contrib/pkg/sets"

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
		trafficTargetOutputs := istio.NewBuilder(ctx, sets.Key(trafficTarget))
		trafficTargetTranslator.Translate(in, trafficTarget, trafficTargetOutputs, reporter)

		if err := t.extensions.PatchTrafficTargetOutputs(ctx, trafficTarget, trafficTargetOutputs); err != nil {
			contextutils.LoggerFrom(ctx).Errorf("failed to apply extension patches for traffic target %v", sets.Key(trafficTarget))
		}

		istioOutputs.Merge(trafficTargetOutputs)
	}

	for _, workload := range in.Workloads().List() {
		workloadOutputs := istio.NewBuilder(ctx, sets.Key(workload))

		// TODO(ilackarms): add translation for workloads when a feature requires us to do so
		//workloadTranslator.Translate(in, workload, workloadOutputs, reporter)

		if err := t.extensions.PatchWorkloadOutputs(ctx, workload, workloadOutputs); err != nil {
			contextutils.LoggerFrom(ctx).Errorf("failed to apply extension patches for traffic target %v", sets.Key(workload))
		}

		istioOutputs.Merge(workloadOutputs)
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
		meshOutputs := istio.NewBuilder(ctx, sets.Key(mesh))
		meshTranslator.Translate(in, mesh, meshOutputs, localOutputs, reporter)

		if err := t.extensions.PatchMeshOutputs(ctx, mesh, meshOutputs); err != nil {
			contextutils.LoggerFrom(ctx).Errorf("failed to apply extension patches for traffic target %v", sets.Key(mesh))
		}

		istioOutputs.Merge(meshOutputs)
	}

	t.totalTranslates++
}
