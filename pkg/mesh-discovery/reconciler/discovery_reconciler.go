package reconciler

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/input"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/smh/pkg/mesh-discovery/translator"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type discoveryReconciler struct {
	ctx          context.Context
	builder      input.Builder
	translator   translator.Translator
	masterClient client.Client
	events       chan struct{}
}

func Start(
	ctx context.Context,
	builder input.Builder,
	translator translator.Translator,
	masterClient client.Client,
	clusters multicluster.ClusterWatcher,
) {

	d := &discoveryReconciler{
		ctx:          ctx,
		builder:      builder,
		translator:   translator,
		masterClient: masterClient,
		events:       make(chan struct{}, 1),
	}

	input.RegisterMultiClusterReconciler(ctx, clusters, d.pushEvent)

	go d.reconcileEventsForever()
}

// simply push a generic event on a reconcile
func (d *discoveryReconciler) pushEvent(_ metav1.Object) error {
	select {
	case d.events <- struct{}{}:
	default:
		// an event is already pending, dropping event is safe
	}
	return nil
}

// reconcile events forever
// blocking (run from a goroutine)
func (d *discoveryReconciler) reconcileEventsForever() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-d.events:
			if err := d.reconcileEvent(); err != nil {
				contextutils.LoggerFrom(d.ctx).Errorw("encountered error reconciling state; retrying", "error", err)

				_ = d.pushEvent(nil)
			}
		}
	}
}

// TODO(ilackarms): it would be nice to make inputSnap and outputSnap available on
// a admin interface, i.e. in JSON format similar to Envoy config dump.
func (d *discoveryReconciler) reconcileEvent() error {
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
