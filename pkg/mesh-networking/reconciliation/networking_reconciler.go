package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/utils/stats"
	"github.com/solo-io/skv2/pkg/predicate"
	"github.com/solo-io/skv2/pkg/reconcile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/mtls"
	"github.com/solo-io/skv2/contrib/pkg/output/errhandlers"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	corev1 "k8s.io/api/core/v1"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/apply"
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
	applier            apply.Applier
	reporter           reporting.Reporter
	translator         translation.Translator
	mgmtClient         client.Client
	multiClusterClient multicluster.Client
	history            *stats.SnapshotHistory
	totalReconciles    int
	verboseMode        bool
}

func Start(
	ctx context.Context,
	builder input.Builder,
	applier apply.Applier,
	reporter reporting.Reporter,
	translator translation.Translator,
	multiClusterClient multicluster.Client,
	mgr manager.Manager,
	history *stats.SnapshotHistory,
	verboseMode bool,
) error {
	d := &networkingReconciler{
		ctx:                ctx,
		builder:            builder,
		applier:            applier,
		reporter:           reporter,
		translator:         translator,
		mgmtClient:         mgr.GetClient(),
		multiClusterClient: multiClusterClient,
		history:            history,
		verboseMode:        verboseMode,
	}

	filterNetworkingEvents := predicate.SimplePredicate{
		Filter: predicate.SimpleEventFilterFunc(isIgnoredSecret),
	}

	_, err := input.RegisterSingleClusterReconciler(ctx, mgr, d.reconcile, time.Second/2, reconcile.Options{}, filterNetworkingEvents)
	return err
}

// reconcile global state
func (r *networkingReconciler) reconcile(obj ezkube.ResourceId) (bool, error) {
	contextutils.LoggerFrom(r.ctx).Debugf("object triggered resync: %T<%v>", obj, sets.Key(obj))

	r.totalReconciles++

	ctx := contextutils.WithLogger(r.ctx, fmt.Sprintf("reconcile-%v", r.totalReconciles))

	inputSnap, err := r.builder.BuildSnapshot(ctx, "mesh-networking", input.BuildOptions{
		// only look at kube clusters in our own namespace
		KubernetesClusters: input.ResourceBuildOptions{
			ListOptions: []client.ListOption{client.InNamespace(defaults.GetPodNamespace())},
		},
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	r.applier.Apply(ctx, inputSnap)

	var errs error

	if err := r.applyTranslation(ctx, inputSnap); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := inputSnap.SyncStatuses(ctx, r.mgmtClient); err != nil {
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

	outputSnap.Apply(ctx, r.mgmtClient, r.multiClusterClient, errHandler)

	r.history.SetInput(in)
	r.history.SetOutput(outputSnap)

	return errHandler.Errors()
}

// returns true if the passed object is a secret which is of a type that is ignored by SMH
func isIgnoredSecret(obj metav1.Object) bool {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return false
	}
	return !mtls.IsSigningCert(secret)
}
