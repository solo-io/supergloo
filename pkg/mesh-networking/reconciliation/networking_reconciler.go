package reconciliation

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/smh/pkg/common/defaults"
	"github.com/solo-io/smh/pkg/mesh-networking/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio"
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
func (d *networkingReconciler) reconcile(_ ezkube.ResourceId) (bool, error) {
	inputSnap, err := d.builder.BuildSnapshot(d.ctx, "mesh-networking", input.BuildOptions{
		// only look at kube clusters in our own namespace
		KubernetesClusters: []client.ListOption{client.InNamespace(defaults.GetPodNamespace())},
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	d.validator.Validate(d.ctx, inputSnap)

	var errs error

	if err := d.syncIstio(inputSnap); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := inputSnap.SyncStatuses(d.ctx, d.masterClient); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

func (d *networkingReconciler) syncIstio(in input.Snapshot) error {
	istioSnap, err := d.istioTranslator.Translate(d.ctx, in, d.reporter)
	if err != nil {
		// internal translator errors should never happen
		return err
	}

	istioSnap.ApplyMultiCluster(d.ctx, d.multiClusterClient, output.ErrorHandlerFuncs{
		HandleWriteErrorFunc: func(resource ezkube.Object, err error) {

		},
		HandleDeleteErrorFunc: nil,
		HandleListErrorFunc:   nil,
	})
}

type writeReporter struct {
}
