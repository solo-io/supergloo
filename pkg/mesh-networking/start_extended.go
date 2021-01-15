package mesh_networking

import (
	"context"
	"time"

	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/pkg/ezkube"

	certissuerinput "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input"
	certissuerreconciliation "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/appmesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm"
	skinput "github.com/solo-io/skv2/contrib/pkg/input"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reconciliation"
)

// Options for extending the functionality of the Networking controller
type ExtensionOpts struct {

	NetworkingReconciler NetworkingReconcilerExtensionOpts

	CertIssuerReconciler CertIssuerReconcilerExtensionOpts
}

type MakeExtensionOpts func(parameters bootstrap.StartParameters) ExtensionOpts

func (opts *ExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {
	opts.NetworkingReconciler.initDefaults(parameters)
	opts.CertIssuerReconciler.initDefaults(parameters)
}

// Options for overriding functionality of the Networking Reconciler
type NetworkingReconcilerExtensionOpts struct {

	// Hook to override how the Cert Issuer Reconciler is registered (defaults to the multi cluster manager)
	RegisterNetworkingReconciler reconciliation.RegisterReconcilerFunc

	// Hook to override the User Snapshot Builder used by Networking Reconciler
	MakeUserSnapshotBuilder func(params bootstrap.StartParameters) input.RemoteBuilder

	// Hook to override Translator used by Networking Reconciler
	MakeTranslator func(translator translation.Translator) translation.Translator

	// Hook to override how the Networking Reconciler applies output snapshots
	SyncNetworkingOutputs reconciliation.SyncOutputsFunc
}

func (opts *NetworkingReconcilerExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {
	if opts.RegisterNetworkingReconciler == nil {
		// use default translator
		opts.RegisterNetworkingReconciler = func(ctx context.Context, reconcile skinput.SingleClusterReconcileFunc, reconcileOpts input.ReconcileOptions) (skinput.InputReconciler, error) {
			return input.RegisterInputReconciler(
				ctx,
				parameters.Clusters,
				func(id ezkube.ClusterResourceId) (bool, error) {
					return reconcile(id)
				},
				parameters.MasterManager,
				reconcile,
				reconcileOpts,
			)
		}
	}
	if opts.MakeTranslator == nil {
		// use default translator
		opts.MakeTranslator = func(translator translation.Translator) translation.Translator {
			return translator
		}
	}
	if opts.SyncNetworkingOutputs == nil {
		// sync outputs to multicluster clients (default)
		opts.SyncNetworkingOutputs = func(
			ctx context.Context,
			outputSnap translation.OutputSnapshots,
			errHandler output.ErrorHandler,
		) {
			outputSnap.ApplyMultiCluster(ctx, parameters.MasterManager.GetClient(), parameters.McClient, errHandler)
		}
	}
	if opts.MakeUserSnapshotBuilder == nil {
		// read from multicluster clients (default)
		opts.MakeUserSnapshotBuilder = func(_ bootstrap.StartParameters) input.RemoteBuilder {
			return input.NewMultiClusterRemoteBuilder(
				parameters.Clusters,
				parameters.McClient,
			)
		}
	}
}

// Options for overriding functionality of the Cert Issuer
type CertIssuerReconcilerExtensionOpts struct {

	// Hook to override how the Cert Issuer Reconciler is registered (defaults to the multi cluster manager)
	RegisterCertIssuerReconciler certissuerreconciliation.RegisterReconcilerFunc

	// Hook to override the Cert Issuer Snapshot Builder used by Cert Issuer Reconciler
	MakeCertIssuerSnapshotBuilder func(params bootstrap.StartParameters) certissuerinput.Builder

	// Hook to override how the Cert Issuer Reconciler syncs the status of inputs (CertificateRequests)
	SyncCertificateIssuerInputStatuses certissuerreconciliation.SyncStatusFunc
}

func (opts *CertIssuerReconcilerExtensionOpts) initDefaults(parameters bootstrap.StartParameters) {
	if opts.MakeCertIssuerSnapshotBuilder == nil {
		// read from multicluster clients (default)
		opts.MakeCertIssuerSnapshotBuilder = func(_ bootstrap.StartParameters) certissuerinput.Builder {
			return certissuerinput.NewMultiClusterBuilder(
				parameters.Clusters,
				parameters.McClient,
			)
		}
	}
	if opts.RegisterCertIssuerReconciler == nil {
		// initialize cert issuer with multicluster clients (default)
		opts.RegisterCertIssuerReconciler = func(
			ctx context.Context,
			reconcile skinput.MultiClusterReconcileFunc,
			reconcileInterval time.Duration,
		) {
			certissuerinput.RegisterMultiClusterReconciler(
				ctx,
				parameters.Clusters,
				reconcile,
				reconcileInterval,
				certissuerinput.ReconcileOptions{},
			)

		}
	}
	if opts.SyncCertificateIssuerInputStatuses == nil {
		// sync statuses to multicluster clients (default)
		opts.SyncCertificateIssuerInputStatuses = func(ctx context.Context, snapshot certissuerinput.Snapshot) error {
			return snapshot.SyncStatusesMultiCluster(ctx, parameters.McClient, certissuerinput.SyncStatusOptions{
				CertificateRequest: true,
			})
		}
	}
}

// custom entryoint for the
// disableMultiCluster - disable multi cluster manager and clientset from being initialized
func StartExtended(ctx context.Context, opts *NetworkingOpts, makeExtensions MakeExtensionOpts, disableMultiCluster bool) error {
	starter := networkingStarter{
		NetworkingOpts: opts,
		makeExtensions: makeExtensions,
	}
	return bootstrap.Start(
		ctx,
		"networking",
		starter.startReconciler,
		*opts.Options,
		disableMultiCluster,
	)
}

type networkingStarter struct {
	*NetworkingOpts

	// callback to configure extensions
	makeExtensions MakeExtensionOpts
}

// start the main reconcile loop
func (s networkingStarter) startReconciler(parameters bootstrap.StartParameters) error {
	extensionOpts := s.makeExtensions(parameters)
	extensionOpts.initDefaults(parameters)

	startCertIssuer(
		parameters.Ctx,
		extensionOpts.CertIssuerReconciler.RegisterCertIssuerReconciler,
		extensionOpts.CertIssuerReconciler.MakeCertIssuerSnapshotBuilder(parameters),
		extensionOpts.CertIssuerReconciler.SyncCertificateIssuerInputStatuses,
		parameters.MasterManager,
	)

	extensionClientset := extensions.NewClientset(parameters.Ctx)

	inputSnapshotBuilder := input.NewSingleClusterLocalBuilder(parameters.MasterManager)

	// contains output resource types read from all registered clusters
	userProvidedSnapshotBuilder := extensionOpts.NetworkingReconciler.MakeUserSnapshotBuilder(parameters)

	reporter := reporting.NewPanickingReporter(parameters.Ctx)


	translator := extensionOpts.NetworkingReconciler.MakeTranslator(translation.NewTranslator(
		istio.NewIstioTranslator(extensionClientset),
		appmesh.NewAppmeshTranslator(),
		osm.NewOSMTranslator(),
	))
	validatingTranslator := extensionOpts.NetworkingReconciler.MakeTranslator(translation.NewTranslator(
		istio.NewIstioTranslator(nil), // the applier should not call the extender
		appmesh.NewAppmeshTranslator(),
		osm.NewOSMTranslator(),
	))

	applier := apply.NewApplier(validatingTranslator)

	return reconciliation.Start(
		parameters.Ctx,
		inputSnapshotBuilder,
		userProvidedSnapshotBuilder,
		applier,
		reporter,
		translator,
		extensionOpts.NetworkingReconciler.RegisterNetworkingReconciler,
		extensionOpts.NetworkingReconciler.SyncNetworkingOutputs,
		parameters.MasterManager.GetClient(),
		parameters.SnapshotHistory,
		parameters.VerboseMode,
		&parameters.SettingsRef,
		extensionClientset,
		s.disallowIntersectingConfig,
		s.watchOutputTypes,
	)
}

func startCertIssuer(
	ctx context.Context,
	registerReconciler certissuerreconciliation.RegisterReconcilerFunc,
	builder certissuerinput.Builder,
	syncInputStatuses certissuerreconciliation.SyncStatusFunc,
	masterManager manager.Manager,
) {
	certissuerreconciliation.Start(
		ctx,
		registerReconciler,
		builder,
		syncInputStatuses,
		masterManager.GetClient(),
	)
}