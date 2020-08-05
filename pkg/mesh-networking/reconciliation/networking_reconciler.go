package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/utils/errhandlers"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/approval"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type networkingReconciler struct {
	ctx                context.Context
	builder            input.Builder
	approver           approval.Approver
	reporter           reporting.Reporter
	translator         translation.Translator
	masterClient       client.Client
	multiClusterClient multicluster.Client
	totalReconciles    int
}

func Start(
	ctx context.Context,
	builder input.Builder,
	validator approval.Approver,
	reporter reporting.Reporter,
	translator translation.Translator,
	multiClusterClient multicluster.Client,
	mgr manager.Manager,
) error {
	d := &networkingReconciler{
		ctx:                ctx,
		builder:            builder,
		approver:           validator,
		reporter:           reporter,
		translator:         translator,
		masterClient:       mgr.GetClient(),
		multiClusterClient: multiClusterClient,
	}

	return input.RegisterSingleClusterReconciler(ctx, mgr, d.reconcile, time.Second/2)
}

// reconcile global state
func (r *networkingReconciler) reconcile(_ ezkube.ResourceId) (bool, error) {
	r.totalReconciles++
	//return false, nil //noop
	ctx := contextutils.WithLogger(r.ctx, fmt.Sprintf("reconcile-%v", r.totalReconciles))
	inputSnap, err := r.builder.BuildSnapshot(ctx, "mesh-networking", input.BuildOptions{
		// only look at kube clusters in our own namespace
		KubernetesClusters: []client.ListOption{client.InNamespace(defaults.GetPodNamespace())},
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	r.approver.Approve(ctx, inputSnap)

	var errs error

	if err := r.applyTranslation(ctx, inputSnap); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := inputSnap.SyncStatuses(ctx, r.masterClient); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

func (r *networkingReconciler) applyTranslation(ctx context.Context, in input.Snapshot) error {
	outputSnap, err := r.translator.Translate(ctx, in, r.reporter)
	if err != nil {
		// internal translator errors should never happen
		return err
	}

	errHandler := errhandlers.AppendingErrHandler{}

	outputSnap.ApplyMultiCluster(ctx, r.multiClusterClient, errHandler)

	outputSnap.ApplyLocalCluster(ctx, r.masterClient, errHandler)

	return errHandler.Errors()
}
