package reconciliation

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/validation"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type networkingReconciler struct {
	ctx                context.Context
	builder            input.Builder
	validator          validation.Validator
	reporter           reporter.Reporter
	istioTranslator    istio.Translator
	masterClient       client.Client
	multiClusterClient multicluster.Client
}

func Start(
	ctx context.Context,
	builder input.Builder,
	validator validation.Validator,
	reporter reporter.Reporter,
	istioTranslator istio.Translator,
	multiClusterClient multicluster.Client,
	mgr manager.Manager,
) error {
	d := &networkingReconciler{
		ctx:                ctx,
		builder:            builder,
		validator:          validator,
		reporter:           reporter,
		istioTranslator:    istioTranslator,
		masterClient:       mgr.GetClient(),
		multiClusterClient: multiClusterClient,
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

	d.validator.Validate(d.ctx, inputSnap)

	var errs error

	if err := d.syncIstio(inputSnap); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := inputSnap.SyncStatuses(d.ctx, d.masterClient); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs
}

func (d *networkingReconciler) syncIstio(translatorSnapshot input.Snapshot) error {
	istioSnap, err := d.istioTranslator.Translate(translatorSnapshot, d.reporter)
	if err != nil {
		// internal translator errors should never happen
		return err
	}

	return istioSnap.ApplyMultiCluster(d.ctx, d.multiClusterClient)
}
