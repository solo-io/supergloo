package reconciliation

import (
	"context"
	"time"

	"github.com/solo-io/service-mesh-hub/pkg/common/utils/stats"
	"github.com/solo-io/skv2/pkg/predicate"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type discoveryReconciler struct {
	ctx          context.Context
	builder      input.Builder
	translator   translation.Translator
	masterClient client.Client
	history      *stats.SnapshotHistory
}

func Start(
	ctx context.Context,
	builder input.Builder,
	translator translation.Translator,
	masterClient client.Client,
	clusters multicluster.ClusterWatcher,
	history *stats.SnapshotHistory,
) {

	r := &discoveryReconciler{
		ctx:          ctx,
		builder:      builder,
		translator:   translator,
		masterClient: masterClient,
		history:      history,
	}

	filterDiscoveryEvents := predicate.SimplePredicate{
		Filter: predicate.SimpleEventFilterFunc(isLeaderElectionObject),
	}

	input.RegisterMultiClusterReconciler(ctx, clusters, r.reconcile, time.Second/2, input.ReconcileOptions{}, filterDiscoveryEvents)
}

func (r *discoveryReconciler) reconcile(obj ezkube.ClusterResourceId) (bool, error) {
	contextutils.LoggerFrom(r.ctx).Debugf("object triggered resync: %T<%v>", obj, sets.Key(obj))

	inputSnap, err := r.builder.BuildSnapshot(r.ctx, "mesh-discovery", input.BuildOptions{})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	outputSnap, err := r.translator.Translate(r.ctx, inputSnap)
	if err != nil {
		// internal translator errors should never happen
		return false, err
	}

	var errs error
	outputSnap.ApplyLocalCluster(r.ctx, r.masterClient, output.ErrorHandlerFuncs{
		HandleWriteErrorFunc: func(resource ezkube.Object, err error) {
			errs = multierror.Append(errs, eris.Wrapf(err, "writing resource %v failed", sets.Key(resource)))
		},
		HandleDeleteErrorFunc: func(resource ezkube.Object, err error) {
			errs = multierror.Append(errs, eris.Wrapf(err, "deleting resource %v failed", sets.Key(resource)))
		},
		HandleListErrorFunc: func(err error) {
			errs = multierror.Append(errs, eris.Wrapf(err, "listing failed"))
		},
	})

	r.history.SetInput(inputSnap)
	r.history.SetOutput(outputSnap)

	return false, errs
}

// returns true if the passed object is used for leader election
func isLeaderElectionObject(obj metav1.Object) bool {
	_, isLeaderElectionObj := obj.GetAnnotations()["control-plane.alpha.kubernetes.io/leader"]
	return isLeaderElectionObj
}
