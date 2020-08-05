package reconciliation

import (
	"context"
	"time"

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
	}

	input.RegisterMultiClusterReconciler(ctx, clusters, r.reconcile, time.Second/2)
}

// TODO(ilackarms): it would be nice to make inputSnap and outputSnap available on
// a admin interface, i.e. in JSON format similar to Envoy config dump.
func (r *discoveryReconciler) reconcile(obj ezkube.ClusterResourceId) (bool, error) {
	if isLeaderElectionObject(obj) {
		contextutils.LoggerFrom(r.ctx).Debugf("ignoring object %v which is being used for leader election", sets.Key(obj))
		return false, nil
	}

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

// returns true if the passed object is used for leader election
func isLeaderElectionObject(obj ezkube.ClusterResourceId) bool {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return false
	}
	_, isLeaderElectionObj := metaObj.GetAnnotations()["control-plane.alpha.kubernetes.io/leader"]
	return isLeaderElectionObj
}
