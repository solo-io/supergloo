package mesh_networking

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/common/schemes"

	certissuerinput "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input"
	certissuerreconciliation "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/appmesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/skv2/pkg/bootstrap"
	"github.com/spf13/pflag"
)

type NetworkingOpts struct {
	*bootstrap.Options
	DisallowIntersectingConfig bool
	WatchOutputTypes           bool
}

func (opts *NetworkingOpts) AddToFlags(flags *pflag.FlagSet) {
	opts.Options.AddToFlags(flags)
	flags.BoolVar(&opts.DisallowIntersectingConfig, "disallow-intersecting-config", false, "if true, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh")
	flags.BoolVar(&opts.WatchOutputTypes, "watch-output-types", true, "if true, Gloo Mesh will watch for the service mesh config output by Gloo Mesh, and resync upon changes.")
}

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts *NetworkingOpts) error {
	return bootstrap.Start(
		ctx,
		StartFunc(opts, func(_ context.Context, _ bootstrap.StartParameters) ExtensionOpts {
			return ExtensionOpts{}
		}),
		*opts.Options,
		schemes.SchemeBuilder,
		false,
	)
}

// custom entryoint for the Networking Reconciler. Used to allow running a customized/extended version of the Networking Reconciler.
// disableMultiCluster - disable multi cluster manager and clientset from being initialized
func StartFunc(opts *NetworkingOpts, makeExtensions MakeExtensionOpts) bootstrap.StartFunc {
	starter := networkingStarter{
		NetworkingOpts: opts,
		makeExtensions: makeExtensions,
	}
	return starter.startReconciler
}

type networkingStarter struct {
	*NetworkingOpts

	// callback to configure extensions
	makeExtensions MakeExtensionOpts
}

// start the main reconcile loop
func (s networkingStarter) startReconciler(ctx context.Context, parameters bootstrap.StartParameters) error {
	extensionOpts := s.makeExtensions(ctx, parameters)
	extensionOpts.initDefaults(parameters)

	if err := startCertIssuer(
		ctx,
		extensionOpts.CertIssuerReconciler.RegisterCertIssuerReconciler,
		extensionOpts.CertIssuerReconciler.MakeCertIssuerSnapshotBuilder(parameters),
		extensionOpts.CertIssuerReconciler.SyncCertificateIssuerInputStatuses,
		parameters.MasterManager,
	); err != nil {
		return err
	}

	extensionClientset := extensions.NewClientset(ctx)

	inputSnapshotBuilder := input.NewSingleClusterLocalBuilder(parameters.MasterManager)

	// contains output resource types read from all registered clusters
	userProvidedSnapshotBuilder := extensionOpts.NetworkingReconciler.MakeUserSnapshotBuilder(parameters)

	reporter := reporting.NewPanickingReporter(ctx)

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
		ctx,
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
		s.DisallowIntersectingConfig,
		s.WatchOutputTypes,
	)
}

func startCertIssuer(
	ctx context.Context,
	registerReconciler certissuerreconciliation.RegisterReconcilerFunc,
	builder certissuerinput.Builder,
	syncInputStatuses certissuerreconciliation.SyncStatusFunc,
	masterManager manager.Manager,
) error {
	return certissuerreconciliation.Start(
		ctx,
		registerReconciler,
		builder,
		syncInputStatuses,
		masterManager.GetClient(),
	)
}
