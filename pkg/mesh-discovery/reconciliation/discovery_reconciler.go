package reconciliation

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/contrib/pkg/sets"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
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

	r := &discoveryReconciler{
		ctx:          ctx,
		builder:      builder,
		translator:   translator,
		masterClient: masterClient,
		events:       workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	input.RegisterMultiClusterReconciler(ctx, clusters, r.reconcile)
}

// TODO(ilackarms): it would be nice to make inputSnap and outputSnap available on
// a admin interface, i.e. in JSON format similar to Envoy config dump.
func (r *discoveryReconciler) reconcile(_ ezkube.ClusterResourceId) (bool, error) {
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

	return false, errs
}
