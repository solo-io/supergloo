package traffictarget

import (
	"context"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/traffictarget/detector"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./traffictarget_translator.go -destination mocks/traffictarget_translator.go

// the traffic-target translator converts deployments with injected sidecars into TrafficTarget CRs
type Translator interface {
	TranslateTrafficTargets(
		services corev1sets.ServiceSet,
		workloads v1alpha2sets.WorkloadSet,
		meshes v1alpha2sets.MeshSet,
	) v1alpha2sets.TrafficTargetSet
}

type translator struct {
	ctx                   context.Context
	trafficTargetDetector detector.TrafficTargetDetector
}

func NewTranslator(ctx context.Context, trafficTargetDetector detector.TrafficTargetDetector) Translator {
	return &translator{ctx: ctx, trafficTargetDetector: trafficTargetDetector}
}

func (t *translator) TranslateTrafficTargets(
	services corev1sets.ServiceSet,
	workloads v1alpha2sets.WorkloadSet,
	meshes v1alpha2sets.MeshSet,
) v1alpha2sets.TrafficTargetSet {

	trafficTargetSet := v1alpha2sets.NewTrafficTargetSet()

	for _, service := range services.List() {
		trafficTarget := t.trafficTargetDetector.DetectTrafficTarget(service, workloads, meshes)
		if trafficTarget == nil {
			continue
		}
		contextutils.LoggerFrom(t.ctx).Debugf("detected traffic target %v", sets.Key(trafficTarget))
		trafficTargetSet.Insert(trafficTarget)
	}
	return trafficTargetSet
}
