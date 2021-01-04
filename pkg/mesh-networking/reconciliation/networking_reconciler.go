package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/skv2/pkg/verifier"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/extensions/v1alpha1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
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
	skv2predicate "github.com/solo-io/skv2/pkg/predicate"
	"github.com/solo-io/skv2/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type networkingReconciler struct {
	ctx                        context.Context
	localBuilder               input.LocalBuilder
	remoteBuilder              input.RemoteBuilder
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
	reconciler                 skinput.InputReconciler
	remoteResourceVerifier     verifier.ServerResourceVerifier
	disallowIntersectingConfig bool
}

// pushNotificationId is a special identifier for a reconcile event triggered by an extension server pushing a notification
var pushNotificationId = &v1.ObjectRef{
	Name: "push-notification-event",
}

func Start(
	ctx context.Context,
	localBuilder input.LocalBuilder,
	remoteBuilder input.RemoteBuilder,
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
	disallowIntersectingConfig bool,
	watchOutputTypes bool,
) error {
	//verifier := verifier.NewVerifier(ctx, map[schema.GroupVersionKind]verifier.ServerVerifyOption{
	//	// only warn (avoids error) if appmesh VirtualNode resource is not available on cluster
	//	schema.GroupVersionKind{
	//		Group:   appmeshv1beta2.GroupVersion.Group,
	//		Version: appmeshv1beta2.GroupVersion.Version,
	//		Kind:    "VirtualNode",
	//	}: verifier.ServerVerifyOption_WarnIfNotPresent,
	//	// only warn (avoids error) if appmesh VirtualMesh resource is not available on cluster
	//	schema.GroupVersionKind{
	//		Group:   appmeshv1beta2.GroupVersion.Group,
	//		Version: appmeshv1beta2.GroupVersion.Version,
	//		Kind:    "VirtualService",
	//	}: verifier.ServerVerifyOption_WarnIfNotPresent,
	//})

	remoteResourceVerifier := buildRemoteResourceVerifier(ctx)

	r := &networkingReconciler{
		ctx:                        ctx,
		localBuilder:               localBuilder,
		remoteBuilder:              remoteBuilder,
		applier:                    applier,
		reporter:                   reporter,
		translator:                 translator,
		mgmtClient:                 mgr.GetClient(),
		multiClusterClient:         multiClusterClient,
		history:                    history,
		verboseMode:                verboseMode,
		settingsRef:                settingsRef,
		extensionClients:           extensionClients,
		disallowIntersectingConfig: disallowIntersectingConfig,
		remoteResourceVerifier:     remoteResourceVerifier,
	}

	// TODO extend skv2 snapshots with singleton object utilities
	// Needed in order to use field selector on metadata.name for Settings CRD.
	if err := mgr.GetFieldIndexer().IndexField(ctx, &settingsv1alpha2.Settings{}, "metadata.name", func(object runtime.Object) []string {
		settings := object.(*settingsv1alpha2.Settings)
		return []string{settings.Name}
	}); err != nil {
		return err
	}

	// watch local input types for changes
	// also watch istio output types for changes, including objects managed by Gloo Mesh itself
	// this should eventually reach a steady state since Gloo Mesh performs equality checks before updating existing objects
	remoteReconcileOptions := reconcile.Options{Verifier: remoteResourceVerifier}
	remoteReconcileOpts := input.RemoteReconcileOptions{
		CertificatesMeshGlooSoloIov1Alpha2IssuedCertificates:  remoteReconcileOptions,
		CertificatesMeshGlooSoloIov1Alpha2PodBounceDirectives: remoteReconcileOptions,
		XdsAgentEnterpriseMeshGlooSoloIov1Alpha1XdsConfigs:    remoteReconcileOptions,
		NetworkingIstioIov1Alpha3DestinationRules:             remoteReconcileOptions,
		NetworkingIstioIov1Alpha3EnvoyFilters:                 remoteReconcileOptions,
		NetworkingIstioIov1Alpha3Gateways:                     remoteReconcileOptions,
		NetworkingIstioIov1Alpha3ServiceEntries:               remoteReconcileOptions,
		NetworkingIstioIov1Alpha3VirtualServices:              remoteReconcileOptions,
		SecurityIstioIov1Beta1AuthorizationPolicies:           remoteReconcileOptions,
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

	reconciler, err := input.RegisterInputReconciler(
		ctx,
		clusters,
		func(id ezkube.ClusterResourceId) (bool, error) {
			return r.reconcile(id)
		},
		mgr,
		r.reconcile,
		input.ReconcileOptions{
			Local: input.LocalReconcileOptions{
				Predicates: []predicate.Predicate{
					skv2predicate.SimplePredicate{
						Filter: skv2predicate.SimpleEventFilterFunc(isIgnoredSecret),
					},
				},
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
		MulticlusterSoloIov1Alpha1KubernetesClusters: input.ResourceLocalBuildOptions{
			ListOptions: []client.ListOption{client.InNamespace(defaults.GetPodNamespace())},
		},
		// ignore NoKindMatchError for AppMesh Mesh CRs
		// (only clusters with AppMesh Controller installed will
		// have these kinds registered)
		//VirtualServices: input.ResourceBuildOptions{
		//	Verifier: r.verifier,
		//},
		//NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes: input.ResourceBuildOptions{
		//	Verifier: r.verifier,
		//},
		SettingsMeshGlooSoloIov1Alpha2Settings: input.ResourceLocalBuildOptions{
			// Ensure that only declared SettingsMeshGlooSoloIov1Alpha2Settings object exists in snapshot.
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
	var userSupplied input.RemoteSnapshot
	//if r.disallowIntersectingConfig {
	//selector := labels.NewSelector()
	//for k := range metautils.TranslatedObjectLabels() {
	//	// select objects without the translated object label key
	//	requirement, err := labels.NewRequirement(k, selection.DoesNotExist, nil)
	//	if err != nil {
	//		// shouldn't happen
	//		return false, err
	//	}
	//	selector = selector.Add([]labels.Requirement{*requirement}...)
	//}
	resourceBuildOptions := input.ResourceRemoteBuildOptions{
		//ListOptions: []client.ListOption{
		//	&client.ListOptions{LabelSelector: selector},
		//},
		Verifier: r.remoteResourceVerifier,
	}
	userSupplied, err = r.remoteBuilder.BuildSnapshot(ctx, "mesh-networking-istio-inputs", input.RemoteBuildOptions{
		CertificatesMeshGlooSoloIov1Alpha2IssuedCertificates:  resourceBuildOptions,
		CertificatesMeshGlooSoloIov1Alpha2PodBounceDirectives: resourceBuildOptions,
		XdsAgentEnterpriseMeshGlooSoloIov1Alpha1XdsConfigs:    resourceBuildOptions,
		NetworkingIstioIov1Alpha3DestinationRules:             resourceBuildOptions,
		NetworkingIstioIov1Alpha3EnvoyFilters:                 resourceBuildOptions,
		NetworkingIstioIov1Alpha3Gateways:                     resourceBuildOptions,
		NetworkingIstioIov1Alpha3ServiceEntries:               resourceBuildOptions,
		NetworkingIstioIov1Alpha3VirtualServices:              resourceBuildOptions,
		SecurityIstioIov1Beta1AuthorizationPolicies:           resourceBuildOptions,
	})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}
	//}

	// apply policies to the discovery resources they target
	r.applier.Apply(ctx, inputSnap, userSupplied)

	// append errors as we still want to sync statuses if applying translation fails
	var errs error

	// translate and apply outputs
	if err := r.applyTranslation(ctx, inputSnap, userSupplied); err != nil {
		errs = multierror.Append(errs, err)
	}

	// update statuses of input objects
	if err := inputSnap.SyncStatuses(ctx, r.mgmtClient, input.LocalSyncStatusOptions{
		SettingsMeshGlooSoloIov1Alpha2Settings:          true,
		DiscoveryMeshGlooSoloIov1Alpha2TrafficTarget:    true,
		DiscoveryMeshGlooSoloIov1Alpha2Workload:         true,
		DiscoveryMeshGlooSoloIov1Alpha2Mesh:             true,
		NetworkingMeshGlooSoloIov1Alpha2TrafficPolicy:   true,
		NetworkingMeshGlooSoloIov1Alpha2AccessPolicy:    true,
		NetworkingMeshGlooSoloIov1Alpha2VirtualMesh:     true,
		NetworkingMeshGlooSoloIov1Alpha2FailoverService: true,
	}); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

func (r *networkingReconciler) applyTranslation(ctx context.Context, in input.LocalSnapshot, userSupplied input.RemoteSnapshot) error {
	if err := r.syncSettings(ctx, in); err != nil {
		// fail early if settings failed to sync
		return err
	}

	outputSnap, err := r.translator.Translate(ctx, in, userSupplied, r.reporter)
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
func (r *networkingReconciler) syncSettings(ctx context.Context, in input.LocalSnapshot) error {
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
		_, _ = r.reconciler.ReconcileLocalGeneric(pushNotificationId)
	})
}

// returns true if the passed object is a secret which is of a type that is ignored by ```22GlooMesh
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
	for groupVersion, kinds := range io.NetworkingRemoteInputTypes {
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
