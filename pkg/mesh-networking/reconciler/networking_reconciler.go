package reconciler

import (
	"context"

	"github.com/solo-io/smh/pkg/mesh-networking/translator"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type networkingReconciler struct {
	ctx          context.Context
	builder      input.Builder
	translator   translator.Translator
	masterClient client.Client
}

func Start(
	ctx context.Context,
	builder input.Builder,
	translator translator.Translator,
	masterClient client.Client,
	mgr manager.Manager,
) error {
	d := &networkingReconciler{
		ctx:          ctx,
		builder:      builder,
		translator:   translator,
		masterClient: masterClient,
	}

	return input.RegisterSingleClusterReconciler(ctx, mgr, d.reconcile)
}

// reconcile global state
func (d *networkingReconciler) reconcile() error {
	inputSnap, err := d.builder.BuildSnapshot(d.ctx, "mesh-networking")
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
