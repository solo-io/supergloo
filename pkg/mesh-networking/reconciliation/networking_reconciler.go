package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/extensions/v1alpha1"
	istioinputs "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input/istio"
	input "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input/networking"
	networkingv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/common/utils/stats"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/snapshotutils"
	"github.com/solo-io/go-utils/contextutils"
	skinput "github.com/solo-io/skv2/contrib/pkg/input"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/predicate"
	"github.com/solo-io/skv2/pkg/reconcile"
	"github.com/solo-io/skv2/pkg/verifier"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type networkingReconciler struct {
	ctx                        context.Context
	builder                    input.Builder
	applier                    apply.Applier
	reporter                   reporting.Reporter
	translator                 translation.Translator
	mgmtClient                 client.Client
	multiClusterClient         multicluster.Client
	history                    *stats.SnapshotHistory
	totalReconciles            int
	verboseMode                bool
	settingsRef                v1.ObjectRef
	extensionClients           extensions.Clientset
	reconciler                 skinput.SingleClusterReconciler
	istioInputs                istioinputs.Builder
	istioResourceVerifier      verifier.ServerResourceVerifier
	disallowIntersectingConfig bool
	watchOutputTypes           bool
}

// pushNotificationId is a special identifier for a reconcile event triggered by an extension server pushing a notification
var pushNotificationId = &v1.ObjectRef{
	Name: "push-notification-event",
}

func Start(
	ctx context.Context,
	builder input.Builder,
	applier apply.Applier,
	reporter reporting.Reporter,
	translator translation.Translator,
	clusters multicluster.ClusterWatcher,
	multiClusterClient multicluster.Client,
	mgr manager.Manager,
	history *stats.SnapshotHistory,
	verboseMode bool,
	settingsRef v1.ObjectRef,
	extensionClients extensions.Clientset,
	istioInputs istioinputs.Builder,
	disallowIntersectingConfig bool,
	watchOutputTypes bool,
) error {

	istioResourceVerifier := buildIstioResourceVerifier(ctx)

	r := &networkingReconciler{
		ctx:                        ctx,
		builder:                    builder,
		applier:                    applier,
		reporter:                   reporter,
		translator:                 translator,
		mgmtClient:                 mgr.GetClient(),
		multiClusterClient:         multiClusterClient,
		history:                    history,
		verboseMode:                verboseMode,
		settingsRef:                settingsRef,
		extensionClients:           extensionClients,
		istioInputs:                istioInputs,
		disallowIntersectingConfig: disallowIntersectingConfig,
		istioResourceVerifier:      istioResourceVerifier,
		watchOutputTypes:           watchOutputTypes,
	}

	filterNetworkingEvents := predicate.SimplePredicate{
		Filter: predicate.SimpleEventFilterFunc(isIgnoredSecret),
	}

	// TODO extend skv2 snapshots with singleton object utilities
	// Needed in order to use field selector on metadata.name for Settings CRD.
	if err := mgr.GetFieldIndexer().IndexField(ctx, &settingsv1alpha2.Settings{}, "metadata.name", func(object runtime.Object) []string {
		settings := object.(*settingsv1alpha2.Settings)
		return []string{settings.Name}
	}); err != nil {
		return err
	}

	reconciler, err := input.RegisterSingleClusterReconciler(ctx, mgr, r.reconcile, time.Second/2, reconcile.Options{}, filterNetworkingEvents)
	if err != nil {
		return err
	}

	// watch istio output types for changes, including objects managed by Gloo Mesh itself
	// this should eventually reach a steady state since Gloo Mesh performs equality checks before updating existing objects
	if r.watchOutputTypes {
		istioinputs.RegisterMultiClusterReconciler(
			ctx,
			clusters,
			func(id ezkube.ClusterResourceId) (bool, error) {
				return r.reconcile(id)
			},
			time.Second/2,
			istioinputs.ReconcileOptions{
				IssuedCertificates:    reconcile.Options{Verifier: istioResourceVerifier},
				PodBounceDirectives:   reconcile.Options{Verifier: istioResourceVerifier},
				XdsConfigs:            reconcile.Options{Verifier: istioResourceVerifier},
				DestinationRules:      reconcile.Options{Verifier: istioResourceVerifier},
				EnvoyFilters:          reconcile.Options{Verifier: istioResourceVerifier},
				Gateways:              reconcile.Options{Verifier: istioResourceVerifier},
				ServiceEntries:        reconcile.Options{Verifier: istioResourceVerifier},
				VirtualServices:       reconcile.Options{Verifier: istioResourceVerifier},
				AuthorizationPolicies: reconcile.Options{Verifier: istioResourceVerifier},
			},
		)
	}

	r.reconciler = reconciler

	return nil
}

// reconcile global state
func (r *networkingReconciler) reconcile(obj ezkube.ResourceId) (bool, error) {
	contextutils.LoggerFrom(r.ctx).Debugf("object triggered resync: %T<%v>", obj, sets.Key(obj))

	r.totalReconciles++

	ctx := contextutils.WithLogger(r.ctx, fmt.Sprintf("reconcile-%v", r.totalReconciles))

	// build the input snapshot from the caches
	inputSnap, err := r.builder.BuildSnapshot(ctx, "mesh-networking", input.BuildOptions{
		// only look at kube clusters in our own namespace
		KubernetesClusters: input.ResourceBuildOptions{
			ListOptions: []client.ListOption{client.InNamespace(defaults.GetPodNamespace())},
		},
		Settings: input.ResourceBuildOptions{
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

	// nil istioInputSnap signals to downstream translators that intersecting config should not be detected
	var existingIstioResources istioinputs.Snapshot
	if r.disallowIntersectingConfig {
		selector := labels.NewSelector()
		for k := range metautils.TranslatedObjectLabels() {
			// select objects without the translated object label key
			requirement, err := labels.NewRequirement(k, selection.DoesNotExist, nil)
			if err != nil {
				// shouldn't happen
				return false, err
			}
			selector.Add([]labels.Requirement{*requirement}...)
		}
		resourceBuildOptions := istioinputs.ResourceBuildOptions{
			ListOptions: []client.ListOption{
				&client.ListOptions{LabelSelector: selector},
			},
			Verifier: r.istioResourceVerifier,
		}
		existingIstioResources, err = r.istioInputs.BuildSnapshot(ctx, "mesh-networking-istio-inputs", istioinputs.BuildOptions{
			IssuedCertificates:    resourceBuildOptions,
			PodBounceDirectives:   resourceBuildOptions,
			XdsConfigs:            resourceBuildOptions,
			DestinationRules:      resourceBuildOptions,
			EnvoyFilters:          resourceBuildOptions,
			Gateways:              resourceBuildOptions,
			ServiceEntries:        resourceBuildOptions,
			VirtualServices:       resourceBuildOptions,
			AuthorizationPolicies: resourceBuildOptions,
		})
		if err != nil {
			// failed to read from cache; should never happen
			return false, err
		}
	}

	// apply policies to the discovery resources they target
	r.applier.Apply(ctx, inputSnap, existingIstioResources)

	// append errors as we still want to sync statuses if applying translation fails
	var errs error

	// translate and apply outputs
	if err := r.applyTranslation(ctx, inputSnap, existingIstioResources); err != nil {
		errs = multierror.Append(errs, err)
	}

	// update statuses of input objects
	if err := inputSnap.SyncStatuses(ctx, r.mgmtClient); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

func (r *networkingReconciler) applyTranslation(ctx context.Context, in input.Snapshot, existingIstioResources istioinputs.Snapshot) error {
	if err := r.syncSettings(ctx, in); err != nil {
		// fail early if settings failed to sync
		return err
	}

	outputSnap, err := r.translator.Translate(ctx, in, existingIstioResources, r.reporter)
	if err != nil {
		// internal translator errors should never happen
		return err
	}

	errHandler := newErrHandler(ctx, in)

	outputSnap.Apply(ctx, r.mgmtClient, r.multiClusterClient, errHandler)

	r.history.SetInput(in)
	r.history.SetOutput(outputSnap)

	return errHandler.Errors()
}

// validate and process the settings stored in the input snapshot.
// exactly one should be present.
// processing/validation errors will be reported to the settings status
func (r *networkingReconciler) syncSettings(ctx context.Context, in input.Snapshot) error {
	settings, err := snapshotutils.GetSingletonSettings(ctx, in)
	if err != nil {
		return err
	}

	settings.Status = settingsv1alpha2.SettingsStatus{
		ObservedGeneration: settings.Generation,
		State:              networkingv1alpha2.ApprovalState_ACCEPTED,
	}

	// update configured NetworkExtensionServers for the extension clients which are called inside the translator.
	return r.extensionClients.ConfigureServers(settings.Spec.NetworkingExtensionServers, func(_ *v1alpha1.PushNotification) {
		// ignore error because underlying impl should never error here
		_, _ = r.reconciler.ReconcileGeneric(pushNotificationId)
	})
}

// returns true if the passed object is a secret which is of a type that is ignored by GlooMesh
func isIgnoredSecret(obj metav1.Object) bool {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return false
	}
	return !mtls.IsSigningCert(secret)
}

// build a verifier that ignores NoKindMatch errors for mesh-specific types
// we expect these errors on clusters on which that mesh is not deployed
func buildIstioResourceVerifier(ctx context.Context) verifier.ServerResourceVerifier {
	options := map[schema.GroupVersionKind]verifier.ServerVerifyOption{}
	for groupVersion, kinds := range io.IstioNetworkingOutputTypes.Resources {
		for _, kind := range kinds {
			gvk := schema.GroupVersionKind{
				Group:   groupVersion.Group,
				Version: groupVersion.Version,
				Kind:    kind,
			}
			options[gvk] = verifier.ServerVerifyOption_IgnoreIfNotPresent
		}
	}
	return verifier.NewVerifier(ctx, options)
}
