package reconciliation

import (
	"context"
	"fmt"
	"time"

	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/reconcile"
	"github.com/solo-io/skv2/pkg/verifier"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/common/utils/stats"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	skpredicate "github.com/solo-io/skv2/pkg/predicate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type discoveryReconciler struct {
	ctx                   context.Context
	discoveryInputBuilder input.DiscoveryInputBuilder
	settingsBuilder       input.SettingsBuilder
	translator            translation.Translator
	masterClient          client.Client
	history               *stats.SnapshotHistory
	verboseMode           bool
	settingsRef           v1.ObjectRef
	verifier              verifier.ServerResourceVerifier
	totalReconciles       int
}

func Start(
	ctx context.Context,
	agentCluster string,
	masterMgr manager.Manager,
	clusters multicluster.Interface,
	mcClient multicluster.Client,
	history *stats.SnapshotHistory,
	verboseMode bool,
	settingsRef v1.ObjectRef,
) error {
	settingsBuilder := input.NewSingleClusterSettingsBuilder(masterMgr)

	var discoveryInputBuilder input.DiscoveryInputBuilder
	if clusters != nil {
		// run in master mode; I/O wired up to local and remote clusters
		discoveryInputBuilder = input.NewMultiClusterDiscoveryInputBuilder(clusters, mcClient)
	} else {
		// run in agent mode;  I/O wired up to local cluster only
		if agentCluster == "" {
			return eris.Errorf("must provide --agent-cluster when running in agent mode")
		}
		discoveryInputBuilder = input.NewSingleClusterDiscoveryInputBuilderWithClusterName(masterMgr, agentCluster)
	}

	verifier := verifier.NewVerifier(ctx, map[schema.GroupVersionKind]verifier.ServerVerifyOption{
		// only warn (avoids error) if appmesh Mesh resource is not available on cluster
		schema.GroupVersionKind{
			Group:   appmeshv1beta2.GroupVersion.Group,
			Version: appmeshv1beta2.GroupVersion.Version,
			Kind:    "Mesh",
		}: verifier.ServerVerifyOption_WarnIfNotPresent,
	})
	r := &discoveryReconciler{
		ctx:                   ctx,
		discoveryInputBuilder: discoveryInputBuilder,
		settingsBuilder:       settingsBuilder,
		translator:            translation.NewTranslator(translation.DefaultDependencyFactory),
		masterClient:          masterMgr.GetClient(),
		history:               history,
		verboseMode:           verboseMode,
		verifier:              verifier,
		settingsRef:           settingsRef,
	}

	filterDiscoveryEvents := skpredicate.SimplePredicate{
		Filter: skpredicate.SimpleEventFilterFunc(isLeaderElectionObject),
	}

	// Needed in order to use field selector on metadata.name for Settings CRD.
	if err := masterMgr.GetFieldIndexer().IndexField(ctx, &settingsv1alpha2.Settings{}, "metadata.name", func(object runtime.Object) []string {
		settings := object.(*settingsv1alpha2.Settings)
		return []string{settings.Name}
	}); err != nil {
		return err
	}

	if clusters != nil {
		// running in master mode; our reconciler should watch local and remote resources
		if _, err := input.RegisterInputReconciler(
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
				Local:             input.LocalReconcileOptions{},
				ReconcileInterval: time.Second / 2,
			},
		); err != nil {
			return err
		}
	} else {
		// running in agent mode; our reconciler should watch only local resources
		if _, err := input.RegisterSingleClusterAgentReconciler(
			ctx,
			masterMgr,
			r.reconcileLocal,
			time.Second/2,
			reconcile.Options{
				Verifier: verifier,
			},
			filterDiscoveryEvents,
		); err != nil {
			return err
		}
	}
	return nil
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

	remoteInputSnap, err := r.discoveryInputBuilder.BuildSnapshot(ctx, "mesh-discovery-remote", input.DiscoveryInputBuildOptions{
		// ignore NoKindMatchError for AppMesh Mesh CRs
		// (only clusters with AppMesh Controller installed will
		// have this kind registered)
		Meshes: input.ResourceDiscoveryInputBuildOptions{
			Verifier: r.verifier,
		},
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	localInputSnap, err := r.settingsBuilder.BuildSnapshot(ctx, "mesh-discovery-local", input.SettingsBuildOptions{
		Settings: input.ResourceSettingsBuildOptions{
			// Ensure that only declared Settings object exists in snapshot.
			ListOptions: []client.ListOption{
				client.InNamespace(r.settingsRef.Namespace),
				client.MatchingFields(map[string]string{
					"metadata.name": r.settingsRef.Name,
				}),
			},
		},
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	settings, err := utils.GetSingletonSettings(ctx, localInputSnap)
	if err != nil {
		return false, err
	}

	outputSnap, err := r.translator.Translate(ctx, remoteInputSnap, settings.Spec.GetDiscovery())
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

	r.history.SetInput(remoteInputSnap)
	r.history.SetOutput(outputSnap)

	return false, errs
}

// returns true if the passed object is used for leader election
func isLeaderElectionObject(obj metav1.Object) bool {
	_, isLeaderElectionObj := obj.GetAnnotations()["control-plane.alpha.kubernetes.io/leader"]
	return isLeaderElectionObj
}
