package destination

import (
	"context"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/destination/detector"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./destination_translator.go -destination mocks/destination_translator.go

// the destination translator converts deployments with injected sidecars into Destination CRs
type Translator interface {
	TranslateDestinations(
		ctx context.Context,
		services corev1sets.ServiceSet,
		pods corev1sets.PodSet,
		nodes corev1sets.NodeSet,
		workloads v1.WorkloadSet,
		meshes v1.MeshSet,
		endpoints corev1sets.EndpointsSet,
	) v1.DestinationSet
}

type translator struct {
	ctx                 context.Context
	destinationDetector detector.DestinationDetector
}

func NewTranslator(destinationDetector detector.DestinationDetector) Translator {
	return &translator{destinationDetector: destinationDetector}
}

func (t *translator) TranslateDestinations(
	ctx context.Context,
	services corev1sets.ServiceSet,
	pods corev1sets.PodSet,
	nodes corev1sets.NodeSet,
	workloads v1.WorkloadSet,
	meshes v1.MeshSet,
	endpoints corev1sets.EndpointsSet,
) v1.DestinationSet {

	DestinationSet := v1.NewDestinationSet()

	for _, service := range services.List() {
		destination := t.destinationDetector.DetectDestination(
			ctx,
			service,
			pods,
			nodes,
			workloads,
			meshes,
			endpoints,
		)
		if destination == nil {
			continue
		}
		contextutils.LoggerFrom(t.ctx).Debugf("detected destination %v", sets.Key(destination))
		DestinationSet.Insert(destination)
	}
	return DestinationSet
}
