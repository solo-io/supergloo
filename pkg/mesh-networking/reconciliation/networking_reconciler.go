package reconciliation

import (
	"context"
	"fmt"
	"time"

	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/pkg/stats"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/settingsutils"
	"github.com/solo-io/skv2/contrib/pkg/output"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/extensions/v1beta1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	skinput "github.com/solo-io/skv2/contrib/pkg/input"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	skv2predicate "github.com/solo-io/skv2/pkg/predicate"
	"github.com/solo-io/skv2/pkg/reconcile"
	"github.com/solo-io/skv2/pkg/verifier"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// function which defines how the Networking reconciler should be registered with internal components.
type RegisterReconcilerFunc func(
	ctx context.Context,
	reconcile skinput.SingleClusterReconcileFunc,
	reconcileOpts input.ReconcileOptions,
) (skinput.InputReconciler, error)

// function which defines how the Networking reconciler should apply its output snapshots.
type SyncOutputsFunc func(
	ctx context.Context,
	inputs input.LocalSnapshot,
	outputSnap *translation.Outputs,
	errHandler output.ErrorHandler,
) error

type networkingReconciler struct {
	ctx                        context.Context
	localBuilder               input.LocalBuilder
	remoteBuilder              input.RemoteBuilder
	applier                    apply.Applier
	reporter                   reporting.Reporter
	translator                 translation.Translator
	syncOutputs                SyncOutputsFunc
	mgmtClient                 client.Client
	history                    *stats.SnapshotHistory
	totalReconciles            int
	verboseMode                bool
	settingsRef                *v1.ObjectRef
	extensionClients           extensions.Clientset
	reconciler                 skinput.InputReconciler
	remoteResourceVerifier     verifier.ServerResourceVerifier
	disallowIntersectingConfig bool
}

var (
	// pushNotificationId is a special identifier for a reconcile event triggered by an extension server pushing a notification
	pushNotificationId = &v1.ObjectRef{
		Name: "push-notification-event",
	}

	// predicates use by the networking reconciler.
	// exported for use in Enterprise.
	NetworkingReconcilePredicates = []predicate.Predicate{
		skv2predicate.SimplePredicate{
			Filter: skv2predicate.SimpleEventFilterFunc(isIgnoredSecret),
		},
	}
)

func Start(
	ctx context.Context,
	localBuilder input.LocalBuilder,
	remoteBuilder input.RemoteBuilder,
	applier apply.Applier,
	reporter reporting.Reporter,
	translator translation.Translator,
	registerReconciler RegisterReconcilerFunc,
	syncOutputs SyncOutputsFunc,
	mgmtClient client.Client,
	history *stats.SnapshotHistory,
	verboseMode bool,
	settingsRef *v1.ObjectRef,
	extensionClients extensions.Clientset,
	disallowIntersectingConfig bool,
	watchOutputTypes bool,
) error {

	ctx = contextutils.WithLogger(ctx, "networking-reconciler")

	remoteResourceVerifier := buildRemoteResourceVerifier(ctx)

	r := &networkingReconciler{
		ctx:                        ctx,
		localBuilder:               localBuilder,
		remoteBuilder:              remoteBuilder,
		applier:                    applier,
		reporter:                   reporter,
		translator:                 translator,
		mgmtClient:                 mgmtClient,
		history:                    history,
		verboseMode:                verboseMode,
		syncOutputs:                syncOutputs,
		settingsRef:                settingsRef,
		extensionClients:           extensionClients,
		disallowIntersectingConfig: disallowIntersectingConfig,
		remoteResourceVerifier:     remoteResourceVerifier,
	}

	// watch local input types for changes
	// also watch istio output types for changes, including objects managed by Gloo Mesh itself
	// this should eventually reach a steady state since Gloo Mesh performs equality checks before updating existing objects
	remoteReconcileOptions := reconcile.Options{Verifier: remoteResourceVerifier}
	remoteReconcileOpts := input.RemoteReconcileOptions{
		IssuedCertificates:    remoteReconcileOptions,
		PodBounceDirectives:   remoteReconcileOptions,
		XdsConfigs:            remoteReconcileOptions,
		DestinationRules:      remoteReconcileOptions,
		EnvoyFilters:          remoteReconcileOptions,
		Gateways:              remoteReconcileOptions,
		ServiceEntries:        remoteReconcileOptions,
		VirtualServices:       remoteReconcileOptions,
		AuthorizationPolicies: remoteReconcileOptions,
		Sidecars:              remoteReconcileOptions,
		Predicates: []predicate.Predicate{
			skv2predicate.SimplePredicate{
				Filter: skv2predicate.SimpleEventFilterFunc(isIgnoredConfigMap),
			},
		},
	}
	// ignore all events (i.e. don't reconcile) if not watching output types
	if !watchOutputTypes {
		remoteReconcileOpts.Predicates = append(
			remoteReconcileOpts.Predicates,
			skv2predicate.SimplePredicate{
				Filter: skv2predicate.SimpleEventFilterFunc(
					func(obj metav1.Object) bool {
						return true
					},
				),
			},
		)
	}

	reconciler, err := registerReconciler(
		ctx,
		r.reconcile,
		input.ReconcileOptions{
			Local: input.LocalReconcileOptions{
				Predicates: NetworkingReconcilePredicates,
			},
			Remote:            remoteReconcileOpts,
			ReconcileInterval: time.Second / 2,
		},
	)
	if err != nil {
		return err
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
	inputSnap, err := r.localBuilder.BuildSnapshot(ctx, "mesh-networking", input.LocalBuildOptions{
		// only look at kube clusters in our own namespace
		KubernetesClusters: input.ResourceLocalBuildOptions{
			ListOptions: []client.ListOption{client.InNamespace(defaults.GetPodNamespace())},
		},
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	if err := r.syncSettings(&ctx, inputSnap); err != nil {
		// fail early if settings failed to sync
		return false, err
	}

	// nil istioInputSnap signals to downstream translators that intersecting config should not be detected
	var userSupplied input.RemoteSnapshot
	if r.disallowIntersectingConfig {
		userSupplied, err = r.buildRemoteSnapshot(ctx, "mesh-networking-istio-inputs", true)
		if err != nil {
			// failed to read from cache; should never happen
			return false, err
		}
	}

	remoteConfig, err := r.buildRemoteSnapshot(ctx, "mesh-networking-generated-outputs", false)
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	// apply policies to the discovery resources they target
	r.applier.Apply(ctx, inputSnap, userSupplied, remoteConfig)

	// append errors as we still want to sync statuses if applying translation fails
	var errs error

	// translate and apply outputs
	if err := r.applyTranslation(ctx, inputSnap, userSupplied, remoteConfig); err != nil {
		errs = multierror.Append(errs, err)
	}

	// update statuses of input objects
	if err := inputSnap.SyncStatuses(ctx, r.mgmtClient, input.LocalSyncStatusOptions{
		// keep this list up to date with all networking status outputs
		Settings:           true,
		Destination:        true,
		Workload:           true,
		Mesh:               true,
		TrafficPolicy:      true,
		AccessPolicy:       true,
		VirtualMesh:        true,
		WasmDeployment:     true,
		AccessLogRecord:    true,
		VirtualDestination: true,
		VirtualGateway:     true,
		VirtualHost:        true,
		RouteTable:         true,
		ServiceDependency:  true,
	}); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

func (r *networkingReconciler) buildRemoteSnapshot(
	ctx context.Context,
	name string,
	userSupplied bool,
) (input.RemoteSnapshot, error) {
	var operator selection.Operator
	if userSupplied {
		// select objects without the translated object label key
		operator = selection.DoesNotExist
	} else {
		// select objects with the translated object label key
		operator = selection.Exists
	}
	selector := labels.NewSelector()
	for k := range metautils.TranslatedObjectLabels() {
		requirement, err := labels.NewRequirement(k, operator, nil)
		if err != nil {
			// shouldn't happen
			return nil, err
		}
		selector = selector.Add([]labels.Requirement{*requirement}...)
	}
	resourceBuildOptions := input.ResourceRemoteBuildOptions{
		ListOptions: []client.ListOption{
			&client.ListOptions{LabelSelector: selector},
		},
		Verifier: r.remoteResourceVerifier,
	}
	return r.remoteBuilder.BuildSnapshot(ctx, name, input.RemoteBuildOptions{
		IssuedCertificates:    resourceBuildOptions,
		PodBounceDirectives:   resourceBuildOptions,
		XdsConfigs:            resourceBuildOptions,
		DestinationRules:      resourceBuildOptions,
		EnvoyFilters:          resourceBuildOptions,
		Gateways:              resourceBuildOptions,
		ServiceEntries:        resourceBuildOptions,
		VirtualServices:       resourceBuildOptions,
		AuthorizationPolicies: resourceBuildOptions,
		Sidecars:              resourceBuildOptions,
	})
}

func (r *networkingReconciler) applyTranslation(
	ctx context.Context,
	in input.LocalSnapshot,
	userSupplied input.RemoteSnapshot,
	generated input.RemoteSnapshot,
) error {

	outputSnap, err := r.translator.Translate(ctx, in, userSupplied, generated, r.reporter)
	if err != nil {
		// internal translator errors should never happen
		return err
	}

	r.history.SetInput(in)
	r.history.SetOutput(outputSnap)

	errHandler := newErrHandler(ctx, in)
	if err := r.syncOutputs(ctx, in, outputSnap, errHandler); err != nil {
		return multierror.Append(err, errHandler.Errors())
	}

	return errHandler.Errors()
}

// stores settings inside the context and initiates connections to extension servers.
// processing/validation errors will be reported to the settings status.
func (r *networkingReconciler) syncSettings(ctx *context.Context, in input.LocalSnapshot) error {
	settings, err := in.Settings().Find(r.settingsRef)
	if err != nil {
		return err
	}

	*ctx = settingsutils.ContextWithSettings(*ctx, settings)

	settings.Status = settingsv1.SettingsStatus{
		ObservedGeneration: settings.Generation,
		State:              commonv1.ApprovalState_ACCEPTED,
	}

	// update configured NetworkExtensionServers for the extension clients which are called inside the translator.
	return r.extensionClients.ConfigureServers(settings.Spec.NetworkingExtensionServers, func(_ *v1beta1.PushNotification) {
		// ignore error because underlying impl should never error here
		_, _ = r.reconciler.ReconcileLocalGeneric(pushNotificationId)
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

// returns true if the passed object is a configmap which is of a type that is ignored by GlooMesh
// this is necessary because Istio-controlled configmaps update very frequently
func isIgnoredConfigMap(obj metav1.Object) bool {
	_, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return false
	}
	return !metautils.IsTranslated(obj)
}

// build a verifier that ignores NoKindMatch errors for mesh-specific types
// we expect these errors on clusters on which that mesh is not deployed
func buildRemoteResourceVerifier(ctx context.Context) verifier.ServerResourceVerifier {
	options := map[schema.GroupVersionKind]verifier.ServerVerifyOption{}
	for groupVersion, kinds := range io.IstioNetworkingOutputTypes.Snapshot {
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
