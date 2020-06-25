package reconciler

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/input"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/translation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type discoveryReconciler struct {
	ctx          context.Context
	builder      input.Builder
	translator   translation.Translator
	masterClient client.Client
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
	}

	input.RegisterMultiClusterReconciler(ctx, clusters, d.reconcile)
}

// reconcile global state
func (d *discoveryReconciler) reconcile() error {
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
