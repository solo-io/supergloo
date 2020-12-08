package reconciliation

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/solo-io/skv2/pkg/reconcile"
	"github.com/solo-io/skv2/pkg/verifier"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/solo-io/gloo-mesh/pkg/common/utils/stats"
	skpredicate "github.com/solo-io/skv2/pkg/predicate"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type discoveryReconciler struct {
	ctx             context.Context
	builder         input.RemoteBuilder
	translator      translation.Translator
	masterClient    client.Client
	managers        multicluster.ManagerSet
	history         *stats.SnapshotHistory
	verboseMode     bool
	verifier        verifier.ServerResourceVerifier
	totalReconciles int
}

func Start(
	ctx context.Context,
	builder input.RemoteBuilder,
	translator translation.Translator,
	masterMgr manager.Manager,
	clusters multicluster.ClusterWatcher,
	history *stats.SnapshotHistory,
	verboseMode bool,
) {
	verifier := verifier.NewVerifier(ctx, map[schema.GroupVersionKind]verifier.ServerVerifyOption{
		// only warn (avoids error) if appmesh Mesh resource is not available on cluster
		schema.GroupVersionKind{
			Group:   appmeshv1beta2.GroupVersion.Group,
			Version: appmeshv1beta2.GroupVersion.Version,
			Kind:    "Mesh",
		}: verifier.ServerVerifyOption_WarnIfNotPresent,
	})
	r := &discoveryReconciler{
		ctx:          ctx,
		builder:      builder,
		translator:   translator,
		masterClient: masterMgr.GetClient(),
		history:      history,
		verboseMode:  verboseMode,
		verifier:     verifier,
	}

	filterDiscoveryEvents := skpredicate.SimplePredicate{
		Filter: skpredicate.SimpleEventFilterFunc(isLeaderElectionObject),
	}

	input.RegisterInputReconciler(
		ctx,
		clusters,
		r.reconcile,
		masterMgr,
		r.reconcileLocal,
		input.ReconcileOptions{
			Remote: input.RemoteReconcileOptions{
				Meshes: reconcile.Options{
					Verifier: verifier,
				},
				Predicates: []predicate.Predicate{filterDiscoveryEvents},
			},
			ReconcileInterval: time.Second / 2,
		},
	)
}

func (r *discoveryReconciler) reconcileLocal(obj ezkube.ResourceId) (bool, error) {
	clusterObj, ok := obj.(ezkube.ClusterResourceId)
	if !ok {
		contextutils.LoggerFrom(r.ctx).Debugf("ignoring event for non cluster type %T %v", obj, sets.Key(obj))
		return false, nil
	}
	return r.reconcile(clusterObj)
}

func (r *discoveryReconciler) reconcile(obj ezkube.ClusterResourceId) (bool, error) {
	r.totalReconciles++
	ctx := contextutils.WithLogger(r.ctx, fmt.Sprintf("reconcile-%v", r.totalReconciles))

	contextutils.LoggerFrom(ctx).Debugf("object triggered resync: %T<%v>", obj, sets.Key(obj))

	inputSnap, err := r.builder.BuildSnapshot(ctx, "mesh-discovery", input.RemoteBuildOptions{
		// ignore NoKindMatchError for AppMesh Mesh CRs
		// (only clusters with AppMesh Controller installed will
		// have this kind registered)
		Meshes: input.ResourceRemoteBuildOptions{
			Verifier: r.verifier,
		},
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	outputSnap, err := r.translator.Translate(ctx, inputSnap)
	if err != nil {
		// internal translator errors should never happen
		return false, err
	}

	var errs error
	outputSnap.ApplyLocalCluster(ctx, r.masterClient, output.ErrorHandlerFuncs{
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
