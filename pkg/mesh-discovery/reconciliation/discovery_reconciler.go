package reconciliation

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/input"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type discoveryReconciler struct {
	ctx          context.Context
	builder      input.Builder
	translator   translation.Translator
	masterClient client.Client
	events       workqueue.RateLimitingInterface
}

func Start(
	ctx context.Context,
	builder input.Builder,
	translator translation.Translator,
	masterClient client.Client,
	clusters multicluster.ClusterWatcher,
) {

	d := &discoveryReconciler{
		ctx:          ctx,
		builder:      builder,
		translator:   translator,
		masterClient: masterClient,
		events:       workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	input.RegisterMultiClusterReconciler(ctx, clusters, d.reconcile)
}

// TODO(ilackarms): it would be nice to make inputSnap and outputSnap available on
// a admin interface, i.e. in JSON format similar to Envoy config dump.
func (d *discoveryReconciler) reconcile(_ ezkube.ClusterResourceId) error {
	inputSnap, err := d.builder.BuildSnapshot(d.ctx, "mesh-discovery")
	if err != nil {
		// failed to read from cache; should never happen
		return err
	}

	outputSnap, err := d.translator.Translate(d.ctx, inputSnap)
	if err != nil {
		// internal translator errors should never happen
		return err
	}

	return outputSnap.Apply(d.ctx, d.masterClient)
}
