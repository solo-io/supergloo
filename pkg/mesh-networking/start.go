package mesh_networking

import (
	"context"
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
	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/spf13/pflag"
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
	return StartExtended(ctx, opts, func(_ bootstrap.StartParameters) ExtensionOpts {
		return ExtensionOpts{}
	}, false)
}

// custom entryoint for the Networking Reconciler. Used to allow running a customized/extended version of the Networking Reconciler.
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
