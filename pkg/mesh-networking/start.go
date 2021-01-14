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
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reconciliation"
)

type NetworkingOpts struct {
	*bootstrap.Options
	disallowIntersectingConfig bool
	watchOutputTypes           bool
}

func (opts *NetworkingOpts) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&opts.disallowIntersectingConfig, "disallow-intersecting-config", false, "if true, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh")
	flags.BoolVar(&opts.watchOutputTypes, "watch-output-types", true, "if true, Gloo Mesh will resync upon changes to the service mesh config output by Gloo Mesh")
}

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts *NetworkingOpts) error {
	starter := networkingStarter{NetworkingOpts: opts}
	return bootstrap.Start(
		ctx,
		"networking",
		starter.startReconciler,
		*opts.Options,
		false,
	)
}

// Options for extending the functionality of the Networking controller
type ExtensionOpts struct {
	// disable multi cluster read/write/reconcile
	DisableMultiCluster bool

	NetworkingReconciler NetworkingReconcilerExtensionOpts

	CertIssuerReconciler CertIssuerReconcilerExtensionOpts
}

func (opts *ExtensionOpts) InitDefaults(parameters bootstrap.StartParameters) {
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
	MakeTranslator func(extensionClientset extensions.Clientset) translation.Translator

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
		opts.MakeTranslator = func(extensionClientset extensions.Clientset) translation.Translator {
			return translation.NewTranslator(
				istio.NewIstioTranslator(extensionClientset),
				appmesh.NewAppmeshTranslator(),
				osm.NewOSMTranslator(),
			)
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
func StartExtended(ctx context.Context, opts *NetworkingOpts, extensions ExtensionOpts) error {
	starter := networkingStarter{
		NetworkingOpts: opts,
		ExtensionOpts:  extensions,
	}
	return bootstrap.Start(
		ctx,
		"networking",
		starter.startReconciler,
		*opts.Options,
		extensions.DisableMultiCluster,
	)
}

type networkingStarter struct {
	*NetworkingOpts
	ExtensionOpts
}

// start the main reconcile loop
func (s networkingStarter) startReconciler(parameters bootstrap.StartParameters) error {

	startCertIssuer(
		parameters.Ctx,
		s.CertIssuerReconciler.RegisterCertIssuerReconciler,
		s.CertIssuerReconciler.MakeCertIssuerSnapshotBuilder(parameters),
		s.CertIssuerReconciler.SyncCertificateIssuerInputStatuses,
		parameters.MasterManager,
	)

	extensionClientset := extensions.NewClientset(parameters.Ctx)

	inputSnapshotBuilder := input.NewSingleClusterLocalBuilder(parameters.MasterManager)

	// contains output resource types read from all registered clusters
	userProvidedSnapshotBuilder := s.NetworkingReconciler.MakeUserSnapshotBuilder(parameters)

	reporter := reporting.NewPanickingReporter(parameters.Ctx)

	translator := s.NetworkingReconciler.MakeTranslator(extensionClientset)
	validatingTranslator := s.NetworkingReconciler.MakeTranslator(nil) // the applier should not call the extender

	applier := apply.NewApplier(validatingTranslator)

	return reconciliation.Start(
		parameters.Ctx,
		inputSnapshotBuilder,
		userProvidedSnapshotBuilder,
		applier,
		reporter,
		translator,
		s.NetworkingReconciler.RegisterNetworkingReconciler,
		s.NetworkingReconciler.SyncNetworkingOutputs,
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
